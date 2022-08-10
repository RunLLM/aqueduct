import { test, expect } from '@playwright/test';

test('Trigger new workflow', async ({ page }) => {
  await page.goto('localhost:8080');
  await page.setViewportSize({width: 1920, height: 1080})
  await page.fill('text=Please enter an API Key ', process.env.API_KEY);
  await page.locator('text=Get Started').click();
  await page.locator("text=Workflows").first().click();
  await page.locator("text=Demo Customer Sentiment").click();
  await page.locator("text=Run Workflow").click();
  await page.locator('text="Run"').click();
});
