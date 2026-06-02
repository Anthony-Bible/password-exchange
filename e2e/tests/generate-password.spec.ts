import { test, expect } from '@playwright/test';

async function openModalAndGenerate(page: any): Promise<void> {
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
    await page.locator('button[data-bs-target="#passwordGeneratorModal"]').click();
    await expect(page.locator('#passwordGeneratorModal')).toBeVisible();
    // Set slider to 20 before generating
    await page.locator('#modal-password-length').evaluate((el: HTMLInputElement) => {
      el.value = '20';
      el.dispatchEvent(new Event('input', { bubbles: true }));
      el.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await page.locator('#modal-generate-password').click();
    await expect(page.locator('#generated-password-section')).toBeVisible();
    const generated = await page.locator('#generated-password').inputValue();
    expect(generated.length).toBe(20);
  });
});
