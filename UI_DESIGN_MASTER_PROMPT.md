# 🔥 COMPLETE MASTER UI/UX PROMPT – YUKTI SaaS PLATFORM

**Purpose**: Copy-paste this into Figma AI, Galileo AI, UX Pilot, Uizard, or any UI-generation tool

---

## 🎯 Platform Overview

Design a modern, clean, enterprise-grade web-based admin dashboard UI for a multi-tenant SaaS platform named **"YUKTI"**.

YUKTI helps organizations securely onboard their cloud infrastructure using AWS IAM Read-Only role ARN, automatically sync cloud resources, store data in PostgreSQL and Redis, and provide cost optimization, budgeting, inventory, and savings analytics.

The UI must look premium, professional, scalable, and trustworthy, suitable for CTOs, DevOps teams, and Finance teams.

---

## 🎨 Visual Style & Design System

### Colors
- **Primary Color**: Deep Blue `#1E3A8A`
- **Secondary / Accent**: Teal & Purple gradients
- **Background**: Dark mode with subtle gradients

### Typography
- **Font Family**: Inter / Roboto (clean sans-serif)
- **Hierarchy**: Clear heading/body distinction

### UI Style
- **Design System**: Material Design / Ant Design inspired
- **Layout**: Desktop-first, responsive
- **CSS Framework**: Tailwind-style spacing & components

### Components
- **Cards**: Rounded corners, soft shadows
- **Tables**: Zebra striping, hover states
- **Icons**: Minimal line icons
- **Charts**: Clean, labeled, tooltip-enabled
- **Logo**: "YUKTI" circular mandala-style logo in header & login

---

## 🔐 AUTHENTICATION & ONBOARDING FLOW (CONNECTED FLOW)

### 1️⃣ Login Page

**Layout**:
- Centered login card
- Dark gradient background
- YUKTI logo at top
- Tagline above form: **"We know the pain"**

**Form Fields**:
- Email / Username (example: `admin@yukti.com`)
- Password
- Primary Login button

**Footer**:
- Privacy & support links

---

### 2️⃣ OTP Verification Page

**Trigger**: Automatically shown after login

**Content**:
- Title: **"Verify your identity"**
- Subtitle: **"Enter the 6-digit OTP sent to your email"**
- 6-digit OTP input boxes (individual boxes for each digit)
- Countdown timer (e.g., "Resend OTP in 59s")
- **Verify & Continue** button
- Success animation after verification

---

### 3️⃣ Cloud Account Onboarding (Wizard Flow)

#### Step 1: Connect AWS Account

**Page Title**: "Securely Connect Your Cloud Account"

**Input Field**:
- IAM Read-Only Role ARN
- Info tooltip: *"YUKTI uses least-privilege access. No write permissions required."*

**Actions**:
- **Validate Role** button
- Success check icon once validated

---

#### Step 2: Resource Syncing

**Layout**: Animated progress view

**Syncing Status Cards**:
- VPC
- EC2
- RDS
- S3
- EKS
- Billing & Cost Explorer

**Status Badges**:
- Pending (gray)
- Syncing (blue, animated)
- Completed (green, checkmark)

**System Message**:
*"Resources are being synced and securely stored in PostgreSQL and Redis."*

---

#### Step 3: Onboarding Complete

**Content**:
- Large success icon (checkmark in circle)
- Text: **"Your cloud environment has been onboarded successfully."**
- CTA: **Go to Dashboard** button

---

## 📊 MAIN DASHBOARD (LANDING PAGE)

### Layout Structure

**Left Sidebar** (collapsible):
- Dashboard (active)
- Tenants
- Users & Roles
- Billing
- Infrastructure
- Inventory
- Whitelisting
- Settings

**Top Header Bar**:
- Organization selector (dropdown)
- Notifications bell icon
- User profile dropdown
- Logout

**Main Content Area**: Grid-based layout

---

### Dashboard Content

#### 🔹 Summary Cards (Top Row)

5 cards in a row:
1. **Total Cloud Spend** - $12,450
2. **Total Cost Saved** - $425.60
3. **Budget Remaining** - $2,750
4. **Active Resources** - 847
5. **Optimization Opportunities** - 7

---

#### 📈 Analytics Section

**Charts** (2x2 grid):
1. **Monthly Cost Trend** (Line Chart)
   - X-axis: Months (Jan, Feb, Mar...)
   - Y-axis: Cost ($)
   - Smooth line with gradient fill

2. **Savings Over Time** (Bar Chart)
   - X-axis: Months
   - Y-axis: Savings ($)
   - Teal bars

3. **Service-Wise Cost Distribution** (Donut Chart)
   - EC2: 45%
   - RDS: 25%
   - S3: 15%
   - ELB: 8%
   - Other: 7%

4. **Budget vs Actual Spend** (Stacked Bar)
   - Budget (gray)
   - Actual (blue)
   - Overspend (red)

---

#### 📦 Inventory Overview

**Table**: Inventory grouped by service

**Columns**:
- Resource Name
- Service Type (EC2, RDS, S3, EKS)
- Region
- Monthly Cost
- Usage %
- Optimization Status (badge: Good / Warning / Critical)

**Example Data**:
- i-0a046ebb489ff3cd7 | EC2 | us-east-1 | $145.60 | 12% CPU | ⚠️ Idle
- db-prod-mysql | RDS | us-west-2 | $280.00 | 67% | ✅ Good
- customer-data-prod | S3 | us-east-1 | $45.20 | 2.3TB | ✅ Good

---

#### 🧾 Budgeting & Whitelisting

**Whitelisted Services Summary**:
- Card showing count of whitelisted resources
- Non-whitelisted cost alerts (red badge)

**Budget Threshold Indicators**:
- Progress bar: 78% used ($9,750 / $12,500)
- Color coding: Green (<70%), Yellow (70-90%), Red (>90%)

---

## 🏢 TENANT MANAGEMENT

### Tenant List Page

**Page Title**: "Tenants"

**Actions**:
- Search & filter bar
- **Add New Tenant** button (top right)

**Table Columns**:
- Tenant No.
- Name
- Company
- Subscription Plan (Free, Pro, Enterprise)
- Expiry Date
- Status (Active, Expired, Suspended)
- Actions (View / Edit icons)

**Example Data**:
- 1 | Shruti | Acme Corp | Pro | 1-Jan-2026 | Active
- 2 | Sangita | TechStart | Enterprise | 15-Mar-2026 | Active

---

### Tenant Details / Edit Page

**Form Layout**: Two-column form

**Fields**:
- Tenant No. (read-only, grayed out)
- Name (editable)
- Email (editable)
- Company (editable)
- Mobile (editable)
- Expiry Date (date picker)
- Subscription Plan (dropdown: Free, Pro, Enterprise)

**Permissions Section**:
- Checkboxes: Admin, Editor, Reader
- Visual tags for selected permissions

**Actions**:
- **Save** button (primary)
- **Cancel** button (secondary)

---

## 👥 USERS & ROLES

**Page Title**: "Users & Roles"

**Table Columns**:
- Name
- Tenant
- Role (Admin / Editor / Reader)
- Actions (Edit / Delete icons)

**Example Data**:
- Footer.co | Acme Corp | Editor
- Reader.co | TechStart | Reader

**Actions**:
- **Add User** button
- **Update User** modal (opens on edit)

---

## 💳 BILLING DASHBOARD

**Page Title**: "Billing"

### Tenant Billing Table

**Columns**:
- Tenant
- Subscription Plan
- Monthly Bill
- Expires
- Status

**Export Buttons** (top right):
- Export to CSV
- Export to PDF
- Export to XLSX

---

### Payment Gateway Integration

**Section Title**: "Onboard Payment Gateway"

**Form Fields**:
- Gateway Name (example: Paytm Gateway)
- API Key (password field)
- API URL (text field: https://www.paytm.com)
- Test / Live toggle switch

**Actions**:
- **Save Gateway** button

---

## ☁️ INFRASTRUCTURE / VPC CONFIGURATION

**Page Title**: "Infrastructure / VPC Configuration"

### Two-Column Layout

#### Left Column: Private Subnet (/16)

**Components**:
- Kubernetes / EKS
- YUKTI Application
- Authorization Services
- PostgreSQL 15.5 (with sync, currency rate)
- Redis Cache
- ClickHouse (Analytics DB)

#### Right Column: Public Subnet (/16)

**Components**:
- Website
- ALB (x2)
- NLB
- Route 53
- CloudFront
- S3 (Customer Data)
- Squid Proxy
- VPC Endpoints
- Support Bot
- Training Videos

**Visual Style**: Card-based layout with icons for each service

---

## 🧩 SYSTEM ARCHITECTURE DIAGRAM

**Page Title**: "System Architecture"

### Diagram Flow

```
User
  ↓
YUKTI SaaS Platform
  ↓
VPC
  ├── Public Subnet
  │   ├── CloudFront
  │   ├── ALB
  │   ├── Website
  │   └── S3
  │
  └── Private Subnet
      ├── EKS / Kubernetes
      ├── YUKTI App
      ├── PostgreSQL
      ├── Redis
      └── ClickHouse
```

**Visual Style**:
- Clean AWS-style architecture diagram
- Clear arrows showing data flow
- Security boundaries highlighted (dashed lines)
- Color-coded components (public: blue, private: purple)

---

## ✨ UX & INTERACTION DETAILS

### Micro-interactions
- **Hover states**: Subtle elevation on cards
- **Active states**: Darker shade on sidebar items
- **Loading states**: Skeleton loaders for tables/charts
- **Success toasts**: Green notification (top-right)
- **Error toasts**: Red notification with retry button

### Responsive Behavior
- **Desktop**: Full sidebar visible
- **Tablet**: Sidebar collapses to icons only
- **Mobile**: Hamburger menu for sidebar

### Accessibility
- **Consistent spacing**: 8px grid system
- **Color contrast**: WCAG AA compliant
- **Keyboard navigation**: Tab order follows visual flow

---

## 🎯 FINAL OUTPUT REQUIREMENT

### Deliverables

Generate **high-fidelity UI mockups** for all screens listed above.

### User Journey Flow

Screens must be visually connected as a user journey:

```
Login → OTP Verification → AWS Onboarding → Resource Sync → Dashboard
```

### Placeholder Data

Use realistic placeholder data:
- **Names**: Shruti, Sangita
- **Companies**: Acme Corp, TechStart, Footer.co, Reader.co
- **Costs**: $12,450, $425.60, $145.60
- **Resources**: i-0a046ebb489ff3cd7, db-prod-mysql, customer-data-prod

### Priority

**Prioritize the Dashboard Analytics experience** as the core product value.

---

## 🚀 NEXT STEPS

If you need:

1. **🔥 Short Midjourney prompt** - For AI image generation
2. **🎨 Figma auto-layout structure** - Component hierarchy
3. **🧭 UX flow diagram** - Mermaid/Lucidchart format
4. **🧱 Design system tokens** - Colors, spacing, typography JSON

Just ask! 😎

---

## 📝 Technical Implementation Notes

### Current Platform Status
- **Backend**: Go 1.23 (Gin framework)
- **Frontend**: React 18 + TypeScript + Tailwind CSS
- **Database**: PostgreSQL 15 + Redis
- **Analytics**: ClickHouse (future migration)
- **Cloud**: AWS (144403604430)
- **Deployment**: Docker Compose → Kubernetes (EKS)

### Existing Features
- ✅ JWT authentication (24-hour expiry)
- ✅ OTP verification via AWS SES
- ✅ AWS IAM role onboarding
- ✅ Multi-region scanning (16 regions)
- ✅ 77 cost detectors
- ✅ RBAC (4 roles: owner, admin, editor, viewer)
- ✅ Admin portal with impersonation

### Design Alignment
This UI design should align with existing backend APIs and database schema documented in:
- `COMPLETE_USER_FLOW.md`
- `API_DOCUMENTATION.md`
- `PRODUCT_OFFERINGS.md`
- `LICENSE_ACTIVATION_MODEL.md`

---

**End of Master Prompt** 🎉
