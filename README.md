# Temporal DSL Orchestrator

An infrastructure provisioning orchestrator built on [Temporal](https://temporal.io/) that uses a YAML-based Domain Specific Language (DSL) to define workflows. These workflows execute provisioning tasks via customizable **executors** such as **Terraform**, **Infracost**, and **Git**. The system exposes a REST API to submit DSL definitions and orchestrate provisioning using Temporal workflows.

---

## 🚀 Features

- 🧩 **YAML-based DSL** for defining provisioning steps
- ⚙️ **Executor model** to plug in Terraform, Infracost, Git, etc.
- 🔁 **Retry & Ignore via Temporal Signals**
- 🌐 RESTful API endpoint to submit and monitor workflows
- 📦 GitHub Actions integration to trigger workflows on code changes // Future

---

