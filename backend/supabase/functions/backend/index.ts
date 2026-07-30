import "@supabase/functions-js/edge-runtime.d.ts";
import { withSupabase } from "@supabase/server";

import { AccessCodeCodec, isValidAccessCode } from "./access_code.ts";
import type { Database, ScoreRow } from "./database.types.ts";

const FUNCTION_NAME = "backend";
const MAX_DEVICE_IDENTIFIER_LENGTH = 512;
const MAX_SAFE_SCORE = Number.MAX_SAFE_INTEGER;
const DEFAULT_SCORE_LIMIT = 100;
const MAX_SCORE_LIMIT = 500;

let codec: AccessCodeCodec | undefined;

class HttpError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

function jsonResponse(
  payload: unknown,
  status = 200,
  headers: HeadersInit = {},
): Response {
  return Response.json(payload, {
    status,
    headers: {
      "Cache-Control": "no-store",
      ...headers,
    },
  });
}

function getCodec(): AccessCodeCodec {
  if (codec) {
    return codec;
  }

  const pepper = Deno.env.get("ACCESS_CODE_PEPPER");
  if (!pepper) {
    throw new Error("ACCESS_CODE_PEPPER is required");
  }

  codec = new AccessCodeCodec(pepper);
  return codec;
}

async function readJsonObject(req: Request): Promise<Record<string, unknown>> {
  let payload: unknown;
  try {
    payload = await req.json();
  } catch {
    throw new HttpError(400, "INVALID_JSON", "Request body must be valid JSON");
  }

  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    throw new HttpError(
      400,
      "INVALID_JSON",
      "Request body must be a JSON object",
    );
  }

  return payload as Record<string, unknown>;
}

function getDeviceUniqueIdentifier(payload: Record<string, unknown>): string {
  const value = payload.deviceUniqueIdentifier;
  if (typeof value !== "string") {
    throw new HttpError(
      400,
      "INVALID_DEVICE_IDENTIFIER",
      "deviceUniqueIdentifier must be a string",
    );
  }

  const normalized = value.trim();
  if (
    normalized.length === 0 ||
    normalized.length > MAX_DEVICE_IDENTIFIER_LENGTH
  ) {
    throw new HttpError(
      400,
      "INVALID_DEVICE_IDENTIFIER",
      `deviceUniqueIdentifier must contain 1-${MAX_DEVICE_IDENTIFIER_LENGTH} characters`,
    );
  }

  return normalized;
}

function getAccessCode(payload: Record<string, unknown>): string {
  const value = payload.accessCode;
  if (typeof value !== "string" || !isValidAccessCode(value)) {
    throw new HttpError(
      400,
      "INVALID_ACCESS_CODE",
      "accessCode must be a valid 20-digit MOD 11-2 code",
    );
  }
  return value;
}

function getSafeInteger(
  value: unknown,
  field: string,
  minimum: number,
): number {
  if (
    typeof value !== "number" ||
    !Number.isSafeInteger(value) ||
    value < minimum ||
    value > MAX_SAFE_SCORE
  ) {
    throw new HttpError(
      400,
      "INVALID_SCORE",
      `${field} must be a safe integer between ${minimum} and ${MAX_SAFE_SCORE}`,
    );
  }
  return value;
}

function getRoute(pathname: string): string {
  const marker = `/${FUNCTION_NAME}`;
  const markerIndex = pathname.lastIndexOf(marker);
  if (
    markerIndex >= 0 &&
    (markerIndex + marker.length === pathname.length ||
      pathname[markerIndex + marker.length] === "/")
  ) {
    return pathname.slice(markerIndex + marker.length) || "/";
  }
  return pathname;
}

function getIntegerQueryParameter(
  url: URL,
  name: string,
  fallback: number,
  minimum: number,
  maximum: number,
): number {
  const raw = url.searchParams.get(name);
  if (raw === null) {
    return fallback;
  }

  const value = Number(raw);
  if (!Number.isSafeInteger(value) || value < minimum || value > maximum) {
    throw new HttpError(
      400,
      "INVALID_PAGINATION",
      `${name} must be an integer between ${minimum} and ${maximum}`,
    );
  }
  return value;
}

function getAuthenticatedUserId(ctx: {
  authMode: string;
  userClaims?: { id?: string; sub?: string } | null;
}): string {
  const userId = ctx.userClaims?.id ?? ctx.userClaims?.sub;
  if (ctx.authMode !== "user" || !userId) {
    throw new HttpError(
      401,
      "AUTHENTICATION_REQUIRED",
      "A valid Supabase access token is required",
    );
  }
  return userId;
}

function serializeScore(row: ScoreRow) {
  return {
    userId: row.user_id,
    score: row.score,
    maxCombo: row.max_combo,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  };
}

export default {
  fetch: withSupabase<Database>(
    { auth: ["user", "publishable"] },
    async (req, ctx) => {
      const url = new URL(req.url);
      const route = getRoute(url.pathname);

      try {
        if (
          req.method === "POST" &&
          (route === "/register" || route === "/access-code")
        ) {
          const payload = await readJsonObject(req);
          const deviceUniqueIdentifier = getDeviceUniqueIdentifier(payload);
          const accessCodeCodec = getCodec();
          const deviceIdentifierHash = await accessCodeCodec
            .deriveDeviceIdentifierHash(
              deviceUniqueIdentifier,
            );

          const { data: existingAccount, error: existingAccountError } =
            await ctx.supabaseAdmin
              .from("device_accounts")
              .select("access_code_hash,access_code_counter")
              .eq("device_identifier_hash", deviceIdentifierHash)
              .maybeSingle();

          if (existingAccountError) {
            throw existingAccountError;
          }

          if (existingAccount) {
            const accessCode = await accessCodeCodec.deriveAccessCodeAtCounter(
              deviceUniqueIdentifier,
              existingAccount.access_code_counter,
            );
            if (!accessCode) {
              throw new Error("Stored access-code counter is invalid");
            }

            const accessCodeHash = await accessCodeCodec.hashAccessCode(
              accessCode,
            );
            if (accessCodeHash !== existingAccount.access_code_hash) {
              throw new Error(
                "ACCESS_CODE_PEPPER does not match the stored account",
              );
            }

            return jsonResponse({
              data: {
                accessCode,
                created: false,
              },
            });
          }

          let nextCounter = 0;
          for (let attempt = 0; attempt < 32; attempt += 1) {
            const candidate = await accessCodeCodec.deriveAccessCode(
              deviceUniqueIdentifier,
              nextCounter,
            );
            nextCounter = candidate.counter + 1;

            const accessCodeHash = await accessCodeCodec.hashAccessCode(
              candidate.accessCode,
            );
            const internalPassword = await accessCodeCodec
              .deriveInternalAuthPassword(
                candidate.accessCode,
              );
            const loginEmail = `${
              crypto.randomUUID().replaceAll("-", "")
            }@device.invalid`;

            const { data: authData, error: authError } = await ctx.supabaseAdmin
              .auth.admin.createUser({
                email: loginEmail,
                password: internalPassword,
                email_confirm: true,
                user_metadata: {
                  authentication_method: "device_access_code",
                },
              });

            if (authError || !authData.user) {
              console.error("Failed to create Supabase Auth user", authError);
              throw new HttpError(
                502,
                "AUTH_USER_CREATE_FAILED",
                "Could not create the user account",
              );
            }

            const authUserId = authData.user.id;
            const { error: insertError } = await ctx.supabaseAdmin
              .from("device_accounts")
              .insert({
                id: authUserId,
                device_identifier_hash: deviceIdentifierHash,
                access_code_hash: accessCodeHash,
                access_code_counter: candidate.counter,
                login_email: loginEmail,
              });

            if (!insertError) {
              return jsonResponse(
                {
                  data: {
                    accessCode: candidate.accessCode,
                    created: true,
                  },
                },
                201,
              );
            }

            const { error: cleanupError } = await ctx.supabaseAdmin.auth.admin
              .deleteUser(authUserId);
            if (cleanupError) {
              console.error(
                "Failed to clean up an unlinked Supabase Auth user",
                cleanupError,
              );
            }

            if (insertError.code !== "23505") {
              throw insertError;
            }

            const {
              data: concurrentlyCreatedAccount,
              error: concurrentAccountError,
            } = await ctx.supabaseAdmin
              .from("device_accounts")
              .select("access_code_hash,access_code_counter")
              .eq("device_identifier_hash", deviceIdentifierHash)
              .maybeSingle();

            if (concurrentAccountError) {
              throw concurrentAccountError;
            }

            if (concurrentlyCreatedAccount) {
              const accessCode = await accessCodeCodec
                .deriveAccessCodeAtCounter(
                  deviceUniqueIdentifier,
                  concurrentlyCreatedAccount.access_code_counter,
                );
              if (!accessCode) {
                throw new Error("Stored access-code counter is invalid");
              }

              const accessCodeHash = await accessCodeCodec.hashAccessCode(
                accessCode,
              );
              if (
                accessCodeHash !==
                  concurrentlyCreatedAccount.access_code_hash
              ) {
                throw new Error(
                  "ACCESS_CODE_PEPPER does not match the stored account",
                );
              }

              return jsonResponse({
                data: {
                  accessCode,
                  created: false,
                },
              });
            }
          }

          throw new HttpError(
            503,
            "ACCESS_CODE_EXHAUSTED",
            "Could not allocate a unique access code",
          );
        }

        if (req.method === "POST" && route === "/login") {
          const payload = await readJsonObject(req);
          const accessCode = getAccessCode(payload);
          const accessCodeCodec = getCodec();
          const accessCodeHash = await accessCodeCodec.hashAccessCode(
            accessCode,
          );

          const { data: account, error: accountError } = await ctx.supabaseAdmin
            .from("device_accounts")
            .select("id,login_email")
            .eq("access_code_hash", accessCodeHash)
            .maybeSingle();

          if (accountError) {
            throw accountError;
          }
          if (!account) {
            throw new HttpError(
              401,
              "INVALID_CREDENTIALS",
              "The access code is invalid",
            );
          }

          const password = await accessCodeCodec.deriveInternalAuthPassword(
            accessCode,
          );
          const { data: signInData, error: signInError } = await ctx.supabase
            .auth.signInWithPassword({
              email: account.login_email,
              password,
            });

          if (signInError || !signInData.session) {
            console.error("Access-code sign-in failed", signInError);
            throw new HttpError(
              401,
              "INVALID_CREDENTIALS",
              "The access code is invalid",
            );
          }

          const { error: lastLoginError } = await ctx.supabaseAdmin
            .from("device_accounts")
            .update({ last_login_at: new Date().toISOString() })
            .eq("id", account.id);
          if (lastLoginError) {
            console.error("Failed to update last_login_at", lastLoginError);
          }

          const session = signInData.session;
          return jsonResponse({
            data: {
              userId: account.id,
              accessToken: session.access_token,
              refreshToken: session.refresh_token,
              tokenType: session.token_type,
              expiresIn: session.expires_in,
              expiresAt: session.expires_at,
            },
          });
        }

        if (
          req.method === "POST" &&
          (route === "/score" || route === "/scores")
        ) {
          const userId = getAuthenticatedUserId(ctx);
          const payload = await readJsonObject(req);
          const score = getSafeInteger(
            payload.score,
            "score",
            0,
          );
          const maxComboValue = payload.maxCombo ?? payload.max_combo ?? 0;
          const maxCombo = getSafeInteger(maxComboValue, "maxCombo", 0);

          const { data: scoreRow, error: scoreError } = await ctx.supabaseAdmin
            .from("scores")
            .upsert(
              {
                user_id: userId,
                score,
                max_combo: maxCombo,
              },
              { onConflict: "user_id" },
            )
            .select("user_id,score,max_combo,created_at,updated_at")
            .single();

          if (scoreError) {
            throw scoreError;
          }

          return jsonResponse({ data: serializeScore(scoreRow) });
        }

        if (
          req.method === "GET" &&
          (route === "/score" || route === "/scores")
        ) {
          const limit = getIntegerQueryParameter(
            url,
            "limit",
            DEFAULT_SCORE_LIMIT,
            1,
            MAX_SCORE_LIMIT,
          );
          const offset = getIntegerQueryParameter(
            url,
            "offset",
            0,
            0,
            1_000_000,
          );

          const {
            data: scoreRows,
            error: scoresError,
            count,
          } = await ctx.supabase
            .from("leaderboard")
            .select("rank,user_id,score,max_combo,updated_at", {
              count: "exact",
            })
            .order("rank", { ascending: true })
            .order("updated_at", { ascending: true })
            .order("user_id", { ascending: true })
            .range(offset, offset + limit - 1);

          if (scoresError) {
            throw scoresError;
          }

          return jsonResponse({
            data: (scoreRows ?? []).map((row) => ({
              rank: row.rank,
              userId: row.user_id,
              score: row.score,
              maxCombo: row.max_combo,
              updatedAt: row.updated_at,
            })),
            pagination: {
              limit,
              offset,
              total: count ?? 0,
            },
          });
        }

        if (req.method === "GET" && route === "/health") {
          return jsonResponse({ status: "ok" });
        }

        return jsonResponse(
          {
            error: {
              code: "NOT_FOUND",
              message: "Route not found",
            },
          },
          404,
        );
      } catch (error) {
        if (error instanceof HttpError) {
          return jsonResponse(
            {
              error: {
                code: error.code,
                message: error.message,
              },
            },
            error.status,
          );
        }

        console.error("Unhandled backend function error", {
          method: req.method,
          route,
          error,
        });
        return jsonResponse(
          {
            error: {
              code: "INTERNAL_SERVER_ERROR",
              message: "Internal server error",
            },
          },
          500,
        );
      }
    },
  ),
};
