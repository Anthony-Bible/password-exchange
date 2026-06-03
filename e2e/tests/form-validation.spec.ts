import { test, expect } from '@playwright/test';

async function waitForBootstrap(page: any): Promise<void> {
  await page.waitForFunction(() => typeof (window as any).bootstrap !== 'undefined');
}

async function enableEmailToggle(page: any): Promise<void> {
  await page.locator('#enableEmail').evaluate((el: HTMLInputElement) => {
    el.checked = true;
    el.dispatchEvent(new Event('change', { bubbles: true }));
    el.dispatchEvent(new Event('input', { bubbles: true }));
  });
}

test.describe('Form validation', () => {
  test('empty message blocks submission', async ({ page }) => {
    await page.goto('/');
    await waitForBootstrap(page);
    await page.getByRole('button', { name: ' Create Secure Link' }).click();
    await expect(page.locator('#form_message')).toHaveClass(/is-invalid/);
    // Error divs are empty (no text), so height is 0; assert via computed CSS rather than visibility
    await expect(page.locator('#messageError')).toHaveCSS('display', 'block');
  });

  test('invalid email format shows error', async ({ page }) => {
    await page.goto('/');
    await waitForBootstrap(page);
    await enableEmailToggle(page);
    await page.getByRole('textbox', { name: 'Your First Name * Help about' }).fill('Anthony');
    await page.locator('#email').fill('notanemail');
    await page.getByRole('textbox', { name: 'Recipient\'s First Name *' }).fill('Bob');
    await page.getByRole('textbox', { name: 'Recipient\'s Email *' }).fill('bob@example.com');
    await page.getByRole('textbox', { name: 'Verification Question * Help' }).fill('What is 2+2?');
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('SecretMessage');
    // Submit button is disabled when email is enabled until Turnstile resolves
    await page.evaluate(() => { (window as any).onTurnstileSuccess('cf-test-token-bypass'); });
    await page.getByRole('button', { name: ' Create Secure Link' }).click();
    await expect(page.locator('#email')).toHaveClass(/is-invalid/);
    await expect(page.locator('#emailError')).toHaveCSS('display', 'block');
  });

  test('email fields required when notification enabled', async ({ page }) => {
    await page.goto('/');
    await waitForBootstrap(page);
    await enableEmailToggle(page);
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('SecretOnly');
    // Submit button is disabled until Turnstile resolves
    await page.evaluate(() => { (window as any).onTurnstileSuccess('cf-test-token-bypass'); });
    await page.getByRole('button', { name: ' Create Secure Link' }).click();
    await expect(page.locator('#firstnameError')).toHaveCSS('display', 'block');
    await expect(page.locator('#emailError')).toHaveCSS('display', 'block');
    await expect(page.locator('#otherFirstnameError')).toHaveCSS('display', 'block');
    await expect(page.locator('#otherEmailError')).toHaveCSS('display', 'block');
    await expect(page.locator('#colorError')).toHaveCSS('display', 'block');
  });

  test('valid form without email submits successfully', async ({ page }) => {
    await page.goto('/');
    await waitForBootstrap(page);
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('JustASecret');
    await page.getByRole('button', { name: ' Create Secure Link' }).click();
    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible({ timeout: 15000 });
  });
});
