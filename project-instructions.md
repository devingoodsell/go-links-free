# Product Requirements Document (PRD)  
**Product Name:** go/links  
**Author:** [Your Name]  
**Date:** February 15, 2025

---

## 1. Overview

**go/links** is an internal URL aliasing service that allows users to create short, memorable URLs (aliases) to redirect to longer, complex URLs. Designed to enhance navigation to internal resources, this system features a responsive, mobile-friendly web interface for users and a dedicated admin interface for oversight and management. The system enforces strict security with optional OKTA SSO integration (which can be enabled or disabled), supports link expiration policies with clear user notifications and confirmation prompts, and includes comprehensive usage analytics and daily backups for disaster recovery. It is built for AWS EKS deployment.

---

## 2. Problem Statement

Internal users often encounter lengthy, error-prone URLs when navigating to key resources. The go/links system addresses these challenges by:
- Providing an intuitive, responsive web interface for managing and accessing go/links.
- Securing access via a configurable authentication system that supports both OKTA SSO and basic email/password login.
- Enabling link lifecycle management via expiration policies.
- Displaying clear status messages for expired links and prompting users to confirm before proceeding.
- Offering detailed usage analytics for monitoring system performance and link engagement.
- Ensuring system reliability with daily backups for disaster recovery.

---

## 3. Goals and Objectives

**Primary Goals:**
- **User Accessibility:** Deliver a seamless, mobile-responsive web interface for go/link creation, editing, and redirection.
- **Security:** Enforce authentication for all users, with the flexibility to use OKTA SSO or basic email/password login.
- **Link Lifecycle Management:** Implement an expiration policy that clearly indicates expired links, provides the redirect URL, and prompts users to confirm before proceeding.
- **Administration:** Supply a robust admin dashboard with detailed usage analytics and link management capabilities.
- **Reliability:** Support up to 1,000 daily active users (DAU) and handle up to 100 requests per second (RPS), while ensuring daily backups for disaster recovery.

**Secondary Objectives:**
- **Integration:** Provide a well-documented REST API for integration with other internal systems.
- **Scalability:** Design the architecture for cloud-native deployment on AWS EKS, enabling future expansion.

---

## 4. Target Audience

- **Internal Employees:** Those who require easy access to internal resources.
- **IT & Security Teams:** Responsible for system maintenance, monitoring, and ensuring secure access.
- **Administrators:** Manage go/links, monitor usage, adjust expiration policies, and perform audits.

---

## 5. Features & Requirements

### A. Core Features

1. **URL Alias Creation and Management**
   - **Description:** Users can create go/links by providing a custom alias and destination URL.
   - **Requirements:**
     - Provide a secure POST endpoint via the web interface for creating new links.
     - Validate inputs to ensure alias uniqueness and URL validity.
     - Enable updating and deletion of links.
     - Allow users to set an expiration date/time upon creation.
     - Reset usage metrics when an expired go/link is updated.
     - Notify users with warnings when a link is nearing expiration.

2. **Redirection**
   - **Description:** The system redirects users from a go/link to its associated destination URL.
   - **Requirements:**
     - Extract the alias from the URL path.
     - Retrieve the destination URL from the datastore.
     - Enforce expiration policy:
       - For expired links, display a message indicating the link has expired.
       - **UI Behavior:** Prompt the user to confirm before proceeding, while also displaying the original destination URL.
     - Issue an HTTP 301/302 redirect for valid, active links.
     - Return a 404 error for non-existent aliases.

3. **Authentication & Authorization**
   - **Description:** All interactions require authentication to ensure security.
   - **Requirements:**
     - **Configurable Authentication:**  
       - Enable OKTA SSO integration for user authentication.  
       - Provide an option to disable OKTA integration, in which case the system will fall back to a basic email/password login.
     - Implement role-based access control to differentiate between standard users and administrators.
     - Ensure secure session management and token validation.
     - **OKTA Integration:** Standard OKTA SSO integration with no known custom policies at this time when enabled.

4. **Expiration Policy**
   - **Description:** Each go/link includes an expiration date/time to manage outdated links.
   - **Requirements:**
     - Allow users to specify an expiration date/time during link creation.
     - Clearly indicate in the UI when a link has expired.
     - Provide the redirect URL alongside an expiration message.
     - Prompt the user with a confirmation dialogue before proceeding to the destination URL.
     - Automatically reset usage metrics when an expired link is updated.

5. **Usage Analytics**
   - **Description:** Track and report on the performance and usage of the go/links system.
   - **Requirements:**
     - Collect overall system metrics: Daily Active Users (DAU), Monthly Active Users (MAU), and Requests Per Second (RPS).
     - Maintain daily, weekly, and total counters for each go/link.
     - Reset individual link counters when an expired go/link is updated.
     - **Granularity:** No additional breakdown (e.g., by department or user role) is required at this time.
     - Present analytics data via the admin interface with visualizations and export options.

### B. Server Components

1. **HTTP API Server**
   - **Technology:** Built using Go (with `net/http` or a third-party router like Gorilla Mux/Chi).
   - **Requirements:**
     - Support CRUD operations for go/links with secure endpoints.
     - Enforce authentication using either OKTA SSO or basic email/password login, based on configuration.
     - Process link expiration and analytics tracking.
     - Optimize for up to 100 RPS.

2. **URL Mapping Store**
   - **Technology Options:** Relational Database (e.g., PostgreSQL) or key/value store optimized for fast lookups.
   - **Requirements:**
     - Store mappings with metadata including creation time, expiration time, and usage counters.
     - Ensure data integrity and fast read/write operations.
     - **Backup & Recovery:** 
       - Daily backups of the PostgreSQL database.
       - Backup retention period: 7 days.
       - In the event of a complete disaster, the system must be redeployable using the latest daily backup.
       - **MTTR (Mean Time To Recovery):** 1 hour.

3. **Deployment & Infrastructure**
   - **Requirements:**
     - Containerize the application using Docker.
     - Deploy on AWS EKS for scalability, reliability, and ease of management.
     - Integrate daily backup processes to support disaster recovery.

### C. Client Components

1. **Web Interface (Primary)**
   - **Description:** A responsive, user-friendly web application for go/link management.
   - **Requirements:**
     - Must support both desktop and mobile devices with a responsive design.
     - Provide forms for link creation, editing, and deletion.
     - Display a clear list of existing go/links, including their status (active or expired) and expiration details.
     - Integrate secure login via OKTA SSO when enabled, or basic email/password login when OKTA is disabled.
     - Implement the expired link UI behavior as described: display an expiration message with the original destination URL and prompt the user to confirm before proceeding.

2. **Admin Interface**
   - **Description:** A dedicated dashboard for administrators to manage, monitor, and audit go/links.
   - **Requirements:**
     - Display detailed information on all go/links, including usage analytics (daily, weekly, and total counters).
     - Provide search and filtering capabilities.
     - Allow administrators to update or remove links.
     - Support data export and audit logging.
     - Include visualizations for overall system metrics (DAU, MAU, RPS).

3. **REST API**
   - **Description:** A comprehensive, documented API for integration with other internal systems.
   - **Requirements:**
     - Expose secure endpoints for link creation, lookup, update, and deletion.
     - Include detailed API documentation and examples.
     - Enforce authentication based on the chosen method (OKTA SSO or basic email/password).

---

## 6. Technical Architecture

### Server Architecture:
- **Programming Language:** Go  
- **Routing:** `net/http` or third-party router (e.g., Gorilla Mux, Chi)  
- **Database:** SQL (e.g., PostgreSQL) or a high-performance key/value store, with schema support for link metadata including expiration and usage counters  
- **Authentication:**  
  - Configurable: OKTA SSO integration when enabled; basic email/password login when disabled.
  - Standard OKTA integration with no custom policies when enabled.
- **Deployment:**  
  - Containerized with Docker  
  - Deployed on AWS EKS  
  - Daily backup routines for disaster recovery (7-day retention, redeployable from the latest backup, MTTR of 1 hour)  
- **Performance:** Optimized to support up to 100 RPS and 1,000 daily active users

### Client Architecture:
- **Web Interface & Admin Interface:**
  - **Frontend Framework:** Modern JavaScript frameworks such as React, Vue, or Angular  
  - **Design:** Responsive design ensuring usability on both desktop and mobile devices  
  - **Integration:** Consumes backend REST API for all operations  
  - **Authentication:** Integrated with either OKTA SSO (if enabled) or basic email/password login for secure access

---

## 7. Success Metrics

- **Adoption & Usage:**
  - Achieve at least 1K daily active users.
  - Sustain up to 100 requests per second reliably.
- **Performance:**
  - Maintain an average redirect latency within acceptable thresholds.
  - Ensure rapid response times for CRUD operations.
- **User Satisfaction:**
  - Positive feedback on ease of use, mobile responsiveness, and clear expired link notifications with confirmation prompts.
  - Low incidence of errors reported by users.
- **Security:**
  - Zero unauthorized access incidents.
  - Robust authentication—whether via OKTA SSO or basic email/password—and secure session management.
- **Analytics Accuracy:**
  - Accurate tracking and reporting of overall system metrics (DAU, MAU, RPS) and individual link usage.
- **Disaster Recovery:**
  - Successful daily backups with a 7-day retention policy.
  - Ability to redeploy services using the latest backup within an MTTR of 1 hour.

---

## 8. Open Questions

- **Expired Link UI Confirmation:**
  - Confirm that the user will be shown a dialogue prompting, “This link has expired. Do you want to proceed to [destination URL]?” with clear options to cancel or confirm.
- **Analytics Detail:**  
  - Currently, no additional breakdown (e.g., by department or user role) is required.
- **Backup Procedures:**  
  - Finalize the detailed backup process and verify recovery steps to meet the MTTR and retention requirements.
- **Authentication Configuration:**  
  - Confirm configuration options and defaults for enabling/disabling OKTA integration.

---

## 9. Dependencies & Assumptions

- **Dependencies:**
  - AWS infrastructure (EKS and related services)
  - A reliable, scalable datastore (e.g., PostgreSQL)
  - OKTA SSO for authentication (optional; basic email/password login as fallback)
  - Containerization (Docker)
  - Expertise in Go and modern JavaScript frameworks

- **Assumptions:**
  - Users will primarily interact with the system through the web interface.
  - The internal network environment is secure and stable.
  - The current load will not exceed 100 RPS with 1K daily active users in the near term.
  - Daily backups will provide adequate support for disaster recovery.

---

## 10. Future Enhancements

- **Enhanced Analytics:**  
  - More detailed usage metrics including geographic data and click patterns.
- **Advanced Security:**  
  - Consider implementing multi-factor authentication and enhanced role-based access controls.
- **Customization Options:**  
  - Allow users to further customize expiration policies and notifications.
- **Integration Expansion:**  
  - Extend the API and integration points for additional internal tools and systems.

---

This PRD provides a comprehensive blueprint for developing the go/links system, ensuring a secure, responsive, and feature-rich solution built for AWS EKS. The system supports configurable authentication (OKTA SSO or basic email/password), clear expired link confirmation behavior, detailed usage analytics, and robust disaster recovery mechanisms.