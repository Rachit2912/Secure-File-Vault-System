# Secure File Vault — Interview Preparation Pack

Welcome to your complete, code-grounded **Interview Preparation Pack** reverse-engineered directly from the Secure File Vault repository.

This pack is specifically designed for **SDE-1 Backend Engineering Interviews**. Every claim, flow, and architecture explanation is directly verified against the Go backend source code, PostgreSQL migrations, and React frontend implementation.

---

## 📖 Recommended Reading Order & Study Strategy

To prepare systematically, read the documents in the following order:

| Step | File | Purpose | Estimated Time |
| :--- | :--- | :--- | :--- |
| **1** | [`01_PROJECT_RECONSTRUCTION.md`](01_PROJECT_RECONSTRUCTION.md) | **System Purpose & Tech Stack**: Understand what the system actually does, entry points, technology breakdown, and architecture diagram. | 10 mins |
| **2** | [`02_ARCHITECTURE_AND_FLOWS.md`](02_ARCHITECTURE_AND_FLOWS.md) | **End-to-End Request Flows**: Learn the 6 core request flows step-by-step from client to storage to answer whiteboard flow questions. | 15 mins |
| **3** | [`03_CODE_WALKTHROUGH.md`](03_CODE_WALKTHROUGH.md) | **Data Model, APIs & Code Walkthrough**: Deep dive into the 15 most important functions, SQL tables, and API inventory. | 25 mins |
| **4** | [`04_SYSTEM_DESIGN_LEARNING.md`](04_SYSTEM_DESIGN_LEARNING.md) | **System Design Concepts**: Master 10 core system design patterns present in this project using the structured failure/trade-off format. | 25 mins |
| **5** | [`05_FAILURE_AND_SCALING.md`](05_FAILURE_AND_SCALING.md) | **Concurrency, Failures & Scaling**: Learn real race conditions, edge case failures, and how to scale the system from 1K to 1M users. | 20 mins |
| **6** | [`06_INTERVIEW_QA.md`](06_INTERVIEW_QA.md) | **Q&A Bank & Personal Pitches**: Practice 50+ realistic interview questions across 5 levels and memorize your project elevator pitches. | 30 mins |
| **7** | [`07_RESUME_CLAIM_VERIFICATION.md`](07_RESUME_CLAIM_VERIFICATION.md) | **Resume Claim Verification**: Ensure you know what is verified in code vs. unsupported phrasing to avoid trap questions. | 10 mins |
| **8** | [`08_FINAL_CHEATSHEET.md`](08_FINAL_CHEATSHEET.md) | **10-Minute Warm-Up Cheat Sheet**: Read this right before your interview for a super fast refresher! | 10 mins |

---

## 🎯 Quick Rules for Your Interview
1. **Be Honest**: Acknowledge design trade-offs (e.g., missing DB indexes, in-memory rate limiting across instances) and explain how you would improve them today.
2. **Ground Answers in Code**: Reference actual Go packages (`gorilla/mux`, `golang.org/x/time/rate`, `lib/pq`), SQL migrations, and Supabase Storage integration.
3. **Focus on System Design**: Use the failure scenarios in `04_SYSTEM_DESIGN_LEARNING.md` and `05_FAILURE_AND_SCALING.md` to demonstrate strong backend fundamentals.
