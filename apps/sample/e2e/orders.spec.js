import { expect, test } from "@playwright/test";

test("creates an order and displays the persisted result", async ({ page }) => {
  const suffix = Date.now();
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Order console" })).toBeVisible();
  await page.getByTestId("customer-id").fill(`e2e-customer-${suffix}`);
  await page.getByTestId("product-id").fill("growth-plan");
  await page.getByTestId("quantity").fill("2");
  await page.getByTestId("price").fill("1250");
  await page.getByTestId("create-order").click();
  const row = page.getByTestId("order-row").filter({ hasText: `e2e-customer-${suffix}` });
  await expect(row).toBeVisible();
  await expect(row).toContainText("$25.00");
  await page.reload();
  await expect(page.getByTestId("order-row").filter({ hasText: `e2e-customer-${suffix}` })).toContainText("$25.00");
});

test("shows an API failure without losing the form", async ({ page }) => {
  await page.route("**/orders", async (route) => {
    if (route.request().method() === "POST") {
      await route.fulfill({ status: 503, contentType: "application/json", body: JSON.stringify({ error: "service unavailable" }) });
      return;
    }
    await route.continue();
  });
  await page.goto("/");
  await page.getByTestId("customer-id").fill("resilient-customer");
  await page.getByTestId("product-id").fill("sample-product");
  await page.getByTestId("quantity").fill("1");
  await page.getByTestId("price").fill("100");
  await page.getByTestId("create-order").click();
  await expect(page.getByRole("alert")).toContainText("service unavailable");
  await expect(page.getByTestId("customer-id")).toHaveValue("resilient-customer");
});

test("prevents invalid quantities before submission", async ({ page }) => {
  const submittedOrders = [];
  page.on("request", (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/orders") {
      submittedOrders.push(request);
    }
  });

  await page.goto("/");
  await page.getByTestId("customer-id").fill("validation-customer");
  await page.getByTestId("product-id").fill("sample-product");
  const quantity = page.getByTestId("quantity");
  await quantity.fill("0");
  await page.getByTestId("create-order").click();

  await expect.poll(() => quantity.evaluate((input) => input.validity.valid)).toBe(false);
  expect(submittedOrders).toHaveLength(0);
});
