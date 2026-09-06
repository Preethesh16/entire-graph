const { chromium } = require('playwright');
const path = require('path');

const baseURL = 'http://127.0.0.1:5176/';
const output = path.resolve(__dirname, 'raw');
const sleep = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds));

async function recordedContext(browser) {
  return browser.newContext({
    viewport: { width: 1600, height: 900 },
    recordVideo: { dir: output, size: { width: 1600, height: 900 } },
    colorScheme: 'dark',
    reducedMotion: 'no-preference',
  });
}

async function finish(page, context, name) {
  const video = page.video();
  await page.close();
  await context.close();
  await video.saveAs(path.join(output, name));
}

(async () => {
  const browser = await chromium.launch({
    headless: true,
    executablePath: '/home/infinity/.cache/ms-playwright/chromium-1234/chrome-linux64/chrome',
  });

  const leadContext = await recordedContext(browser);
  const leadPage = await leadContext.newPage();
  await leadPage.goto(baseURL, { waitUntil: 'networkidle' });
  await sleep(4500);
  await leadPage.getByLabel('Your name').fill('Preethesh');
  await leadPage.getByLabel('Your role').fill('Engineering lead');
  await sleep(1800);
  await leadPage.getByRole('button', { name: 'Create team room' }).click();
  await leadPage.getByRole('heading', { name: 'Your team is connected.' }).waitFor();
  await sleep(6500);
  const leadSession = JSON.parse(await leadPage.evaluate(() => sessionStorage.getItem('spidey-sense.browser-team.v1')));
  const inviteURL = new URL(baseURL);
  inviteURL.hash = `join?${new URLSearchParams({ team: leadSession.teamId, invite: leadSession.inviteCode })}`;
  await finish(leadPage, leadContext, '01-create-room.webm');

  const joinContext = await recordedContext(browser);
  const joinPage = await joinContext.newPage();
  await joinPage.goto(inviteURL.toString(), { waitUntil: 'networkidle' });
  await sleep(4200);
  await joinPage.getByLabel('Your name').fill('Deepthi');
  await joinPage.getByLabel('Your role').fill('UI engineer');
  await sleep(1600);
  await joinPage.getByRole('button', { name: 'Join team room' }).click();
  await joinPage.getByRole('heading', { name: 'Your team is connected.' }).waitFor();
  await sleep(6000);
  const teammateSession = JSON.parse(await joinPage.evaluate(() => sessionStorage.getItem('spidey-sense.browser-team.v1')));
  await finish(joinPage, joinContext, '02-join-browser.webm');

  const tourContext = await recordedContext(browser);
  await tourContext.addInitScript((session) => sessionStorage.setItem('spidey-sense.browser-team.v1', JSON.stringify(session)), leadSession);
  const tourPage = await tourContext.newPage();
  const keepAlive = setInterval(async () => {
    try {
      await tourContext.request.post(`${baseURL}api/v1/agents/${teammateSession.agentId}/heartbeat`, {
        headers: { Authorization: `Bearer ${teammateSession.agentToken}` },
        data: { mission_id: 'product-surface', blocker: '', last_activity: new Date().toISOString() },
      });
    } catch {}
  }, 10000);
  await tourPage.goto(baseURL, { waitUntil: 'domcontentloaded' });
  await tourPage.getByRole('heading', { name: 'Your team is connected.' }).waitFor();
  await sleep(5000);
  await tourPage.getByRole('button', { name: '2. Plan & assign' }).click();
  await sleep(3500);
  const owner = tourPage.locator('select[aria-label^="Owner for"]').first();
  await owner.selectOption(teammateSession.memberId);
  await sleep(6500);
  await tourPage.getByRole('button', { name: '3. Graph space' }).click();
  await sleep(4000);
  await tourPage.locator('.graph-experience').scrollIntoViewIfNeeded();
  await sleep(11000);
  await tourPage.locator('.scene-node').first().click();
  await sleep(5500);
  await tourPage.getByRole('button', { name: 'Refresh evidence' }).click();
  await tourPage.waitForFunction(() => {
    const button = [...document.querySelectorAll('button')].find((element) => element.textContent?.includes('Refresh evidence'));
    return button && !button.disabled;
  }, undefined, { timeout: 30000 });
  await sleep(8500);
  await tourPage.getByRole('button', { name: 'Live activity' }).click();
  await sleep(8000);
  clearInterval(keepAlive);
  await finish(tourPage, tourContext, '03-plan-graph-activity.webm');

  await browser.close();
  console.log(JSON.stringify({ leadSession: { teamId: leadSession.teamId }, clips: 3 }));
})().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
