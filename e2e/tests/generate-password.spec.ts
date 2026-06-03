import { test, expect } from '@playwright/test';

async function waitForBootstrap(page: any): Promise<void> {
  page.on('response', (r: any) => {
    if (r.url().includes('bootstrap')) console.log('[BST]', r.status(), r.url());
  });
  page.on('requestfailed', (r: any) => {
    if (r.url().includes('bootstrap')) console.error('[BST FAIL]', r.url(), r.failure()?.errorText);
  });
  await page.waitForFunction(() => {
    console.log('[BST CHECK] bootstrap type:', typeof (window as any).bootstrap);
    return typeof (window as any).bootstrap !== 'undefined';
  }).catch(async (e: any) => {
    const title = await page.title();
    console.error('[BST TIMEOUT] page title was:', title);
    throw e;
  });
}

async function openModalAndGenerate(page: any): Promise<void> {
  await waitForBootstrap(page);
  await page.locator('button[data-bs-target="#passwordGeneratorModal"]').click();
  await expect(page.locator('#passwordGeneratorModal')).toBeVisible();
  // The generated-password section is hidden until the user clicks "Generate Password"
  await page.locator('#modal-generate-password').click();
  await expect(page.locator('#generated-password-section')).toBeVisible();
}

test.describe('Generate password modal', () => {
  test('modal opens and generates a password', async ({ page }) => {
    await page.goto('/');
    await openModalAndGenerate(page);
    const generated = await page.locator('#generated-password').inputValue();
    expect(generated.length).toBeGreaterThan(0);
  });

  test('insert copies password into message field and closes modal', async ({ page }) => {
    await page.goto('/');
    await openModalAndGenerate(page);
    await page.locator('#insert-password-btn').click();
    await expect(page.locator('#passwordGeneratorModal')).not.toBeVisible();
    // The password regenerates live so we just verify something non-empty was inserted
    const inserted = await page.locator('#form_message').inputValue();
    expect(inserted.length).toBeGreaterThan(0);
  });

  test('generated password length matches slider value', async ({ page }) => {
    await page.goto('/');
    await waitForBootstrap(page);
    await page.locator('button[data-bs-target="#passwordGeneratorModal"]').click();
    await expect(page.locator('#passwordGeneratorModal')).toBeVisible();
    // Set slider to 20 before generating
    await page.locator('#modal-password-length').evaluate((el: HTMLInputElement) => {
      el.value = '20';
      el.dispatchEvent(new Event('input', { bubbles: true }));
      el.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await expect(page.locator('#modalLengthValue')).toHaveText('20');
    await page.locator('#modal-generate-password').click();
    await expect(page.locator('#generated-password-section')).toBeVisible();
    const generated = await page.locator('#generated-password').inputValue();
    expect(generated.length).toBe(20);
  });
});
