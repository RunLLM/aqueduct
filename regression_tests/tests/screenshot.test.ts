import { test, expect } from '@playwright/test';

test('Screenshot Workflow page', async ({ page, browser, browserName }) => {
  await page.goto('localhost:8080');

  await page.fill('text=Please enter an API Key ', process.env.API_KEY);
  await page.setViewportSize({width: 3000, height: 1500})
  await page.locator('text=Get Started').click();
  await page.locator("text=Workflows").first().click();
  await page.locator("text=Demo Customer Sentiment").click();
  await page.locator("text=Notifications").click();
  await page.screenshot({
    path: `./screenshots/${browserName}-${process.env.imgname}.png` ,
    fullPage: true})
});
