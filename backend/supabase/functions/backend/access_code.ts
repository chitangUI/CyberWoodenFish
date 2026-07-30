const encoder = new TextEncoder();
const ACCESS_CODE_DATA_MODULUS = 10_000_000_000_000_000_000n;
const MAX_DERIVATION_ATTEMPTS = 1024;

function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join(
    "",
  );
}

function bytesToBigInt(bytes: Uint8Array): bigint {
  let result = 0n;
  for (const byte of bytes) {
    result = (result << 8n) | BigInt(byte);
  }
  return result;
}

export function calculateMod112CheckCharacter(data: string): string {
  if (!/^\d+$/.test(data)) {
    throw new TypeError("MOD 11-2 data must contain digits only");
  }

  let total = 0;
  for (const character of data) {
    total = ((total + Number(character)) * 2) % 11;
  }

  const result = (12 - total) % 11;
  return result === 10 ? "X" : String(result);
}

export function isValidAccessCode(accessCode: string): boolean {
  if (!/^\d{20}$/.test(accessCode)) {
    return false;
  }

  const data = accessCode.slice(0, 19);
  return calculateMod112CheckCharacter(data) === accessCode[19];
}

export class AccessCodeCodec {
  readonly #keyPromise: Promise<CryptoKey>;

  constructor(pepper: string) {
    if (pepper.length < 32) {
      throw new Error("ACCESS_CODE_PEPPER must contain at least 32 characters");
    }

    this.#keyPromise = crypto.subtle.importKey(
      "raw",
      encoder.encode(pepper),
      {
        name: "HMAC",
        hash: "SHA-256",
      },
      false,
      ["sign"],
    );
  }

  async #hmac(message: string): Promise<Uint8Array> {
    const key = await this.#keyPromise;
    const signature = await crypto.subtle.sign(
      "HMAC",
      key,
      encoder.encode(message),
    );
    return new Uint8Array(signature);
  }

  async deriveDeviceIdentifierHash(
    deviceUniqueIdentifier: string,
  ): Promise<string> {
    return bytesToHex(
      await this.#hmac(`device-id:v1\0${deviceUniqueIdentifier}`),
    );
  }

  async deriveAccessCodeAtCounter(
    deviceUniqueIdentifier: string,
    counter: number,
  ): Promise<string | null> {
    if (!Number.isSafeInteger(counter) || counter < 0) {
      throw new RangeError(
        "Access code counter must be a non-negative integer",
      );
    }

    const digest = await this.#hmac(
      `access-code:v1\0${counter}\0${deviceUniqueIdentifier}`,
    );
    const numericData = (
      bytesToBigInt(digest) % ACCESS_CODE_DATA_MODULUS
    ).toString().padStart(19, "0");
    const checkCharacter = calculateMod112CheckCharacter(numericData);

    // ISO 7064 MOD 11-2 can produce X. This service promises digits only,
    // so callers advance the deterministic counter when that happens.
    return checkCharacter === "X" ? null : `${numericData}${checkCharacter}`;
  }

  async deriveAccessCode(
    deviceUniqueIdentifier: string,
    startCounter = 0,
  ): Promise<{ accessCode: string; counter: number }> {
    for (
      let counter = startCounter;
      counter < startCounter + MAX_DERIVATION_ATTEMPTS;
      counter += 1
    ) {
      const accessCode = await this.deriveAccessCodeAtCounter(
        deviceUniqueIdentifier,
        counter,
      );
      if (accessCode) {
        return { accessCode, counter };
      }
    }

    throw new Error("Could not derive a numeric MOD 11-2 access code");
  }

  async hashAccessCode(accessCode: string): Promise<string> {
    return bytesToHex(
      await this.#hmac(`access-code-lookup:v1\0${accessCode}`),
    );
  }

  async deriveInternalAuthPassword(accessCode: string): Promise<string> {
    const digest = bytesToHex(
      await this.#hmac(`auth-password:v1\0${accessCode}`),
    );
    return `Aa1!${digest}`;
  }
}
