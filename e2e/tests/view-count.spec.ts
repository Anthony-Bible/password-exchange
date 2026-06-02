import { test, expect } from '@playwright/test';

async function createMessage(page: any, secret: string, maxViewCount?: number): Promise<string> {
  await page.goto('/');
  if (maxViewCount !== undefined) {
    await page.locator('#max_view_count').fill(String(maxViewCount));
  }
  await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill(secret);
  await page.getByRole('button', { name: ' Create Secure Link' }).click();
  await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible({ timeout: 15000 });
  return page.getByRole('alert').getByRole('textbox').inputValue();
}

test.describe('Max view count', () => {
  test('default view count allows 5 views before deletion', async ({ page }) => {
    const shareUrl = await createMessage(page, 'DefaultViewCountSecret');

    // First view — page should decrypt and show the 5-view limit
    await page.goto(shareUrl);
    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#view-count-text')).toContainText('will be deleted after 5 views');
    await expect(page.locator('#message-content-input')).toHaveValue('DefaultViewCountSecret');
  });

  test('default view count self-destructs after 5 views', async ({ page }) => {
    const shareUrl = await createMessage(page, 'FiveViewSecret');

    // Views 1–5: each should decrypt successfully
    for (let i = 1; i <= 5; i++) {
      await page.goto(shareUrl);
      await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
      await expect(page.locator('#message-content-input')).toHaveValue('FiveViewSecret');
      if (i < 5) {
        await page.goto('/'); // navigate away to force full reload on next visit
      }
    }

    // View 5 is the last — should show "reached the maximum view count"
    await expect(page.locator('#view-count-text')).toContainText('reached the maximum view count');

    await page.goto('/');

    // View 6 — permanently gone
    await page.goto(shareUrl);
    await expect(page.getByText('This page self-destructed')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Permanently deleted')).toBeVisible();
  });

  test('custom view count of 2 expires after 2 views', async ({ page }) => {
    const shareUrl = await createMessage(page, 'ShortLivedSecret', 2);

    // First view — should decrypt normally, warn about 1 remaining
    await page.goto(shareUrl);
    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#view-count-text')).toContainText('will be deleted after 2 views');
    await expect(page.locator('#message-content-input')).toHaveValue('ShortLivedSecret');

    // Navigate away so the next goto triggers a full reload (not a fragment-only update)
    await page.goto('/');

    // Second view — hits the limit, message should be marked as deleted
    await page.goto(shareUrl);
    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#view-count-text')).toContainText('reached the maximum view count');

    await page.goto('/');

    // Third view — message is permanently gone, server returns the 404 self-destruct page
    await page.goto(shareUrl);
    await expect(page.getByText('This page self-destructed')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Permanently deleted')).toBeVisible();
  });
});
