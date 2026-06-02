import { test, expect } from '@playwright/test';

test.describe('Submit secure password form', () => {
  test('submits form without email notification and decrypted content matches', async ({ page }) => {
    await page.goto('/');

    const secret = 'SuperSecret123!';
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill(secret);
    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible();

    const linkBox = page.getByRole('alert').getByRole('textbox');
    await expect(linkBox).toHaveValue(/\/decrypt\/.+#key=.+/);

    // Navigate to the generated link and verify the decrypted content matches
    const shareUrl = await linkBox.inputValue();
    await page.goto(shareUrl);

    // Page auto-decrypts when #key= fragment is present — wait for the content to appear
    await expect(page.locator('#decrypted-message')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#message-content-input')).toHaveValue(secret);
  });

  test('generated secure link keeps encryption key in URL fragment', async ({ page }) => {
    await page.goto('/');

    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('AnotherSecret!');
    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible();

    const link = await page.getByRole('alert').getByRole('textbox').inputValue();

    // Key must live in the fragment so it never hits the server
    expect(link).toMatch(/https?:\/\/.+\/decrypt\/.+#key=.+/);
  });

  test('shows optional security fields on load', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByRole('textbox', { name: 'Passphrase (Optional)' })).toBeVisible();
    await expect(page.getByRole('spinbutton', { name: 'Max View Count (Optional)' })).toBeVisible();
    await expect(page.getByRole('spinbutton', { name: 'Expiration (Optional)' })).toBeVisible();
  });

  test('submits form with passphrase and custom view count', async ({ page }) => {
    await page.goto('/');

    await page.getByRole('textbox', { name: 'Passphrase (Optional)' }).fill('mypassphrase');
    await page.getByRole('spinbutton', { name: 'Max View Count (Optional)' }).fill('3');
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('SecretWithOptions!');

    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible();
  });

  test('enabling email notification reveals sender and recipient fields', async ({ page }) => {
    await page.goto('/');

    // Fields should not exist before checking the box
    await expect(page.getByRole('textbox', { name: 'Your First Name * Help about' })).not.toBeVisible();
    await expect(page.getByRole('textbox', { name: 'Recipient\'s Email *' })).not.toBeVisible();

    await page.getByRole('checkbox', { name: 'Send email notification to recipient' }).check();

    await expect(page.getByRole('textbox', { name: 'Your First Name * Help about' })).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'Your Email * Help about your' })).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'Recipient\'s First Name *' })).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'Recipient\'s Email *' })).toBeVisible();
    await expect(page.getByRole('textbox', { name: 'Verification Question * Help' })).toBeVisible();
  });

  test('submits form with email notification enabled', async ({ page }) => {
    // Mock the message creation endpoint — bypasses server-side Turnstile validation
    // since we can't solve a real CAPTCHA in an automated browser.
    await page.route('**/api/v1/messages', async (route) => {
      if (route.request().method() === 'POST') {
        const origin = new URL(route.request().url()).origin;
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            messageId: 'test-message-id-abc123',
            webUrl: `${origin}/decrypt/testUniqueId`,
            isClientEncrypted: false,
            expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
          }),
        });
      } else {
        await route.continue();
      }
    });

    await page.route('**/api/v1/messages/*/notify', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ notificationSent: true }),
      });
    });

    await page.goto('/');

    await page.getByRole('checkbox', { name: 'Send email notification to recipient' }).check();

    await page.getByRole('textbox', { name: 'Your First Name * Help about' }).fill('Anthony');
    await page.getByRole('textbox', { name: 'Your Email * Help about your' }).fill('sender@anthony.bible');
    await page.getByRole('textbox', { name: 'Recipient\'s First Name *' }).fill('Test Recipient');
    await page.getByRole('textbox', { name: 'Recipient\'s Email *' }).fill('recipient@anthony.bible');
    await page.getByRole('textbox', { name: 'Verification Question * Help' }).fill('What is the name of my cat?');
    await page.getByRole('textbox', { name: 'Password or Secret Message *' }).fill('SuperSecret123!');

    // Turnstile cannot be solved in headless automation. Call onTurnstileSuccess()
    // directly — it's the global callback the real widget fires, and it sets the
    // local turnstileToken variable and enables the submit button.
    await page.evaluate(() => {
      (window as any).onTurnstileSuccess('cf-test-token-bypass');
    });

    await page.getByRole('button', { name: ' Create Secure Link' }).click();

    await expect(page.getByRole('heading', { name: 'Secure Link Created Successfully!' })).toBeVisible();
    await expect(page.getByText('Email notification sent successfully!')).toBeVisible();

    const linkBox = page.getByRole('alert').getByRole('textbox');
    await expect(linkBox).toHaveValue(/\/decrypt\/.+/);
  });
});
