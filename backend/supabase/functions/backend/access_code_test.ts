import assert from "node:assert/strict";
import test from "node:test";

import {
  AccessCodeCodec,
  calculateMod112CheckCharacter,
  isValidAccessCode,
} from "./access_code.ts";

const TEST_PEPPER = "test-only-pepper-that-is-at-least-32-characters";

test("calculates known ISO 7064 MOD 11-2 check characters", () => {
  assert.equal(calculateMod112CheckCharacter("000000021825009"), "7");
  assert.equal(calculateMod112CheckCharacter("000000021694233"), "X");
});

test("validates a 20-digit access code", () => {
  const data = "1234567890123456789";
  const checkCharacter = calculateMod112CheckCharacter(data);
  assert.notEqual(checkCharacter, "X");
  assert.equal(isValidAccessCode(`${data}${checkCharacter}`), true);
  assert.equal(isValidAccessCode(`${data}0`), checkCharacter === "0");
  assert.equal(isValidAccessCode("123"), false);
});

test("derives a stable, numeric access code from a device identifier", async () => {
  const codec = new AccessCodeCodec(TEST_PEPPER);
  const first = await codec.deriveAccessCode("device-123");
  const second = await codec.deriveAccessCode("device-123");

  assert.deepEqual(first, second);
  assert.match(first.accessCode, /^\d{20}$/);
  assert.equal(isValidAccessCode(first.accessCode), true);
});

test("separates device hashes, access-code hashes, and auth passwords", async () => {
  const codec = new AccessCodeCodec(TEST_PEPPER);
  const { accessCode } = await codec.deriveAccessCode("device-123");

  const deviceHash = await codec.deriveDeviceIdentifierHash("device-123");
  const accessCodeHash = await codec.hashAccessCode(accessCode);
  const password = await codec.deriveInternalAuthPassword(accessCode);

  assert.match(deviceHash, /^[0-9a-f]{64}$/);
  assert.match(accessCodeHash, /^[0-9a-f]{64}$/);
  assert.notEqual(deviceHash, accessCodeHash);
  assert.match(password, /^Aa1![0-9a-f]{64}$/);
});
