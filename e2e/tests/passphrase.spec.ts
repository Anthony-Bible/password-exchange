import { test, expect } from '@playwright/test';

async function createMessage(page: any, secret: string, passphrase?: string): Promise<string> {
  await page.goto('/');
  if (passphrase !== undefined) {
    await page.locator('#other_lastname').fill(passphrase);
  }
  await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill(secret);
  await page.getByRole('button', { name: ' Create Secure Link' }).click();
  await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible({ timeout: 15000 });
  return page.getByRole('alert').getByRole('textbox').inputValue();
}

test.describe('Passphrase flow', () => {
  test('passphrase required — shows access form', async ({ page }) => {
    const shareUrl = await createMessage(page, 'PassphraseSecret', 'correctpassphrase');
    await page.goto(shareUrl);
    await expect(page.locator('#access-form')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#access-form')).toContainText('Passphrase Required');
  });

  test('wrong passphrase shows error', async ({ page }) => {
    const shareUrl = await createMessage(page, 'PassphraseSecret2', 'correctpassphrase');
    await page.goto(shareUrl);
    await expect(page.locator('#access-form')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-bs-target="#loginModal"]').click();
    await expect(page.locator('#loginModal')).toBeVisible();
    await page.locator('#passphrase').fill('wrongpassphrase');
    await page.locator('#decrypt-btn').click();
    await expect(page.locator('#passphrase-error')).toContainText('Invalid passphrase. Please try again.');
  });

  test('correct passphrase decrypts message', async ({ page }) => {
    const secret = 'PassphraseDecryptSecret';
    const passphrase = 'mysecretpassphrase';
    const shareUrl = await createMessage(page, secret, passphrase);
    await page.goto(shareUrl);
    await expect(page.locator('#access-form')).toBeVisible({ timeout: 10000 });
    await page.locator('[data-bs-target="#loginModal"]').click();
    await expect(page.locator('#loginModal')).toBeVisible();
    await page.locator('#passphrase').fill(passphrase);
    await page.locator('#decrypt-btn').click();
    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#message-content-input')).toHaveValue(secret);
  });
});
