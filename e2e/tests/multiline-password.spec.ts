import { test, expect } from '@playwright/test';

test.describe('Multi-line password preservation', () => {
  test('multi-line password preserves newlines after decrypt', async ({ page }) => {
    await page.goto('/');

    const multilineSecret = 'Line1\nLine2\nLine3';
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill(multilineSecret);
    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible();

    const shareUrl = await page.getByRole('alert').getByRole('textbox').inputValue();
    await page.goto(shareUrl);

    // Page auto-decrypts when #key= fragment is present — wait for content
    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#message-content-input')).toHaveValue(multilineSecret);
  });

  test('credential block with multiple fields preserves structure after decrypt', async ({ page }) => {
    await page.goto('/');

    const credentials = 'hostname: db.example.com\nusername: admin\npassword: P@ssw0rd!\nport: 5432';
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill(credentials);
    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible();

    const shareUrl = await page.getByRole('alert').getByRole('textbox').inputValue();
    await page.goto(shareUrl);

    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#message-content-input')).toHaveValue(credentials);
  });
});
