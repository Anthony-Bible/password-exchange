(function () {
  function isSupported() {
    return !!(window.crypto && window.crypto.subtle);
  }

  function bytesToBase64Url(bytes) {
    const binary = Array.from(bytes, b => String.fromCharCode(b)).join('');
    return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '');
  }

  function base64UrlToBytes(value) {
    const padded = value.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat((4 - (value.length % 4)) % 4);
    const binary = atob(padded);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i);
    return bytes;
  }

  async function generateKey() {
    return crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, true, ['encrypt', 'decrypt']);
  }

  async function exportKeyBase64Url(key) {
    const raw = await crypto.subtle.exportKey('raw', key);
    return bytesToBase64Url(new Uint8Array(raw));
  }

  async function importKeyBase64Url(keyBase64Url) {
    const raw = base64UrlToBytes(keyBase64Url);
    return crypto.subtle.importKey('raw', raw, { name: 'AES-GCM' }, false, ['encrypt', 'decrypt']);
  }

  async function encryptText(plaintext) {
    const key = await generateKey();
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encoded = new TextEncoder().encode(plaintext);
    const encrypted = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, encoded);
    const keyBase64Url = await exportKeyBase64Url(key);
    const ciphertext = `${bytesToBase64Url(iv)}.${bytesToBase64Url(new Uint8Array(encrypted))}`;
    return { ciphertext, keyBase64Url };
  }

  async function decryptText(keyBase64Url, ivCiphertext) {
    const [ivPart, cipherPart] = String(ivCiphertext || '').split('.');
    if (!ivPart || !cipherPart) throw new Error('invalid ciphertext');
    const key = await importKeyBase64Url(keyBase64Url);
    const iv = base64UrlToBytes(ivPart);
    const ciphertext = base64UrlToBytes(cipherPart);
    const decrypted = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext);
    return new TextDecoder().decode(decrypted);
  }

  window.E2EEncrypt = {
    isSupported,
    bytesToBase64Url,
    base64UrlToBytes,
    generateKey,
    exportKeyBase64Url,
    importKeyBase64Url,
    encryptText,
    decryptText,
  };
})();
