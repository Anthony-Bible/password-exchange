import { test, expect, type Page } from '@playwright/test';

// A file ID that will never exist in storage — used to exercise the 404 path
// on the download page without needing to mock the API.
const NONEXISTENT_FILE_ID = 'e2e-nonexistent-file-id-000000000000';

// A syntactically valid base64url key (44 chars, 32 zero bytes).
// Used for download-page UI tests that don't reach the storage layer.
const DUMMY_ENCODED_KEY = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=';

test.describe('File upload flow', () => {
  test('uploads a file alongside a message and shows share URL with file params', async ({ page }) => {
    await page.goto('/');

    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('SecretWithFile!');

    const fileChooserPromise = page.waitForEvent('filechooser');
    await page.locator('#file-upload-input').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles({
      name: 'test-document.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('hello world'),
    });

    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible({ timeout: 15000 });

    // The share URL fragment must contain both file ID and key params
    const linkValue = await page.locator('#secure-url').inputValue();
    expect(linkValue).toMatch(/fid=.+/);
    expect(linkValue).toMatch(/fk=.+/);
  });

  test('shows file name and size after file selection', async ({ page }) => {
    await page.goto('/');

    const fileChooserPromise = page.waitForEvent('filechooser');
    await page.locator('#file-upload-input').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles({
      name: 'my-secret.pdf',
      mimeType: 'application/pdf',
      buffer: Buffer.from('pdf content here'),
    });

    await expect(page.locator('.file-upload-label-text')).toHaveText('my-secret.pdf');
    await expect(page.locator('.file-upload-size-hint')).not.toBeEmpty();
  });

  test('file upload failure is non-fatal — secure link still shown without file params', async ({ page }) => {
    // Force the file initiation endpoint to fail while letting real message creation run.
    await page.route('**/api/v1/files/initiate', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'INTERNAL_ERROR', message: 'storage unavailable' }),
      });
    });

    await page.goto('/');

    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('SecretNoFile!');

    const fileChooserPromise = page.waitForEvent('filechooser');
    await page.locator('#file-upload-input').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles({
      name: 'fail-me.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('will not upload'),
    });

    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    // Message creation succeeded, so the secure link heading appears
    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible({ timeout: 15000 });

    // File params must be absent since the upload failed
    const linkValue = await page.locator('#secure-url').inputValue();
    expect(linkValue).not.toContain('fid=');
    expect(linkValue).not.toContain('fk=');

    // The file upload error banner should be visible
    await expect(page.locator('.file-upload-error')).toBeVisible();
  });
});

test.describe('File download page', () => {
  test('shows download button when URL fragment contains file ID and key', async ({ page }) => {
    await page.goto(`/files/${NONEXISTENT_FILE_ID}#k=${DUMMY_ENCODED_KEY}`);

    await expect(page.getByRole('heading', { name: 'Secure File Download' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Download File' })).toBeVisible();
    await expect(page.locator('#download-error')).not.toBeVisible();
  });

  test('shows error when URL fragment is missing the key', async ({ page }) => {
    await page.goto(`/files/${NONEXISTENT_FILE_ID}`);

    await expect(page.locator('#download-error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#download-error-message')).toContainText(/missing|key/i);
  });

  test('file attached to message can be downloaded from the decrypt page', async ({ page }) => {
    // Step 1: upload a file alongside a message
    await page.goto('/');

    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('RoundTripTest!');

    const fileChooserPromise = page.waitForEvent('filechooser');
    await page.locator('#file-upload-input').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles({
      name: 'roundtrip.txt',
      mimeType: 'text/plain',
      buffer: Buffer.from('round trip file content'),
    });

    await page.getByRole('button', { name: ' Create Secure Link' }).click();
    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible({ timeout: 15000 });

    const shareUrl = await page.locator('#secure-url').inputValue();

    // Step 2: navigate to the decrypt URL — it already carries fid= and fk= in the fragment
    await page.goto(shareUrl);

    // Page auto-decrypts; wait for the decrypted message to appear
    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 15000 });

    // The file section should appear once decryption completes
    await expect(page.locator('#file-section')).toBeVisible();

    // Step 3: click the download button and verify it reaches the "Downloaded" state
    const downloadPromise = page.waitForEvent('download');
    await page.locator('#file-download-btn').click();
    await downloadPromise;

    await expect(page.locator('#file-download-btn')).toHaveClass(/btn-success/, { timeout: 10000 });
    await expect(page.locator('#file-download-status')).toContainText('File saved');
  });

  test('shows error state when file does not exist', async ({ page }) => {
    // NONEXISTENT_FILE_ID will produce a real 404 from the server
    await page.goto(`/files/${NONEXISTENT_FILE_ID}#k=${DUMMY_ENCODED_KEY}`);

    await expect(page.getByRole('button', { name: 'Download File' })).toBeVisible();
    await page.getByRole('button', { name: 'Download File' }).click();

    await expect(page.locator('#download-error')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#download-error-message')).toContainText(/not found/i);
  });
});
