// Simple UI Element Test Script
const puppeteer = require('puppeteer');

async function testUIElements() {
  console.log('🧪 Starting UI Element Verification...\n');
  
  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();
  
  try {
    // Test 1: Homepage
    console.log('📍 Testing Homepage (/)...');
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle0' });
    
    const title = await page.title();
    console.log(`✅ Page Title: ${title}`);
    
    // Check for React root
    const rootExists = await page.$('#root') !== null;
    console.log(`${rootExists ? '✅' : '❌'} React Root Element: ${rootExists}`);
    
    // Test 2: Login Page
    console.log('\n📍 Testing Login Page (/login)...');
    await page.goto('http://localhost:3000/login', { waitUntil: 'networkidle0' });
    
    // Check for common login elements
    const emailField = await page.$('input[type="email"], input[name="email"], input[placeholder*="email" i]') !== null;
    const passwordField = await page.$('input[type="password"]') !== null;
    const submitButton = await page.$('button[type="submit"], button:contains("Sign"), button:contains("Login")') !== null;
    
    console.log(`${emailField ? '✅' : '❌'} Email/Username Field: ${emailField}`);
    console.log(`${passwordField ? '✅' : '❌'} Password Field: ${passwordField}`);
    console.log(`${submitButton ? '✅' : '❌'} Submit Button: ${submitButton}`);
    
    // Test 3: Dashboard Page
    console.log('\n📍 Testing Dashboard Page (/dashboard)...');
    await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle0' });
    
    const dashboardContent = await page.content();
    const hasDashboardText = dashboardContent.includes('Dashboard') || dashboardContent.includes('dashboard');
    console.log(`${hasDashboardText ? '✅' : '❌'} Dashboard Content: ${hasDashboardText}`);
    
    // Test 4: Navigation Elements
    console.log('\n📍 Testing Navigation Elements...');
    const navElements = await page.$$('nav, .nav, [role="navigation"]');
    const hasNavigation = navElements.length > 0;
    console.log(`${hasNavigation ? '✅' : '❌'} Navigation Menu: ${hasNavigation}`);
    
    // Test 5: Check for JavaScript errors
    console.log('\n📍 Checking for JavaScript Errors...');
    const jsErrors = [];
    page.on('pageerror', error => jsErrors.push(error.message));
    
    await page.reload({ waitUntil: 'networkidle0' });
    
    if (jsErrors.length === 0) {
      console.log('✅ No JavaScript Errors');
    } else {
      console.log(`❌ JavaScript Errors Found: ${jsErrors.length}`);
      jsErrors.forEach(error => console.log(`   - ${error}`));
    }
    
  } catch (error) {
    console.log(`❌ Test Error: ${error.message}`);
  } finally {
    await browser.close();
  }
  
  console.log('\n🎯 UI Element Verification Complete!');
}

// Run if puppeteer is available, otherwise skip
(async () => {
  try {
    await testUIElements();
  } catch (error) {
    console.log('⚠️  Puppeteer not available. Running basic HTML content test instead...\n');
    
    // Fallback: Basic HTML content test
    const { exec } = require('child_process');
    
    console.log('📍 Testing HTML Content...');
    
    exec('curl -s http://localhost:3000', (error, stdout) => {
      if (error) {
        console.log(`❌ Homepage Error: ${error.message}`);
        return;
      }
      
      const hasReactRoot = stdout.includes('id="root"');
      const hasTitle = stdout.includes('Yukti');
      const hasMetaDescription = stdout.includes('FinOps');
      
      console.log(`${hasReactRoot ? '✅' : '❌'} React Root: ${hasReactRoot}`);
      console.log(`${hasTitle ? '✅' : '❌'} Yukti Title: ${hasTitle}`);
      console.log(`${hasMetaDescription ? '✅' : '❌'} FinOps Description: ${hasMetaDescription}`);
      
      console.log('\n🎯 Basic HTML Test Complete!');
    });
  }
})();