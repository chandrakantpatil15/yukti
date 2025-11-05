# Yukti FinOps Platform - Development Progress

## Current Focus: Optimized Assessment-Based Architecture

### Key Architectural Decisions Made

**1. Resource Identification Strategy**
- **Primary Key**: ARN (`arn:aws:ec2:region:account:instance/i-xxxxx`) for universal uniqueness
- **Log Correlation Hierarchy**:
  1. Search for ARN in logs (highest confidence)
  2. Search for service-specific ID (`i-0123456789abcdef0`)
  3. Search by IP address (private/public)
  4. Search by hostname/DNS
  5. Search by tag-based identifiers (Application, Service names)

**2. Database Optimization Strategy**
- **Problem**: Heavy time-series data in PostgreSQL, complex customer setup
- **Solution**: Assessment-based ratings system
  - External assessment engine processes time-series data
  - PostgreSQL stores only lightweight ratings/scores
  - Plug-and-play for customers without time-series infrastructure

**3. Assessment Categories (User-Configurable Thresholds)**
- **Underutilized**: CPU < 20%, Memory < 25% (default, user-configurable)
- **Overutilized**: CPU > 80%, Memory > 80% (sustained periods)
- **Intermittent/Bursting**: Occasional spikes, otherwise low usage
- **Batch/Scheduled**: Predictable usage windows (nightly ETL, ML training)
- **Idle**: Minimal usage, candidate for termination

**4. Smart Query Optimization**
- User-selectable timeline queries (avoid full table scans)
- Caching mechanism for live data requirements
- Assessment engine as separate microservice (batch processing)

### Database Schema Evolution

**Current Tables:**
1. `yt_aws_resources` - Resource inventory with ARN as primary identifier
2. `yt_aws_pricing` - AWS pricing data (8,499 records cached)
3. `yt_assessment_config` - User-configurable thresholds per tenant
4. `yt_resource_assessments` - Lightweight assessment results (replaces heavy metrics)
5. `yt_assessment_history` - Historical ratings for trend analysis
6. `yt_assessment_cache` - Caching for live data requirements

### Implementation Status

**✅ Completed:**
- Real AWS data integration (pricing + resources)
- Multi-tenant architecture design
- Resource identification strategy
- Optimized assessment-based schema

**🔄 Current Work:**
- Assessment engine design (separate microservice)
- User-configurable threshold system
- Timeline-based query optimization

**📋 Next Steps:**
1. Build assessment engine microservice
2. Implement log correlation with fallback hierarchy
3. Create user dashboard for threshold configuration
4. Add time-series visualization integration

### Key Benefits of Current Approach

1. **Lightweight PostgreSQL** - Only business logic, no heavy time-series data
2. **Customer-Agnostic** - Works with any existing monitoring setup
3. **Scalable** - Assessment engine can scale independently
4. **Cost-Effective** - Minimal database storage requirements
5. **User-Friendly** - Configurable thresholds, timeline queries, caching for live data

### Technology Stack
- **Backend**: Go with PostgreSQL
- **Assessment Engine**: Separate Go microservice (batch processing)
- **Time-Series**: External (CloudWatch, customer's existing tools)
- **Caching**: PostgreSQL-based with TTL
- **Deployment**: Kubernetes with microservice architecture

### Enterprise SaaS Model
- **Target**: 1000+ enterprise customers
- **Cost Structure**: $1/customer/month infrastructure cost
- **Security**: ReadOnly AWS permissions with tenant isolation
- **Scalability**: Microservice architecture with independent scaling