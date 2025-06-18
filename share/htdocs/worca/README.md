# 🚀 Worca - Workforce Orchestrator

> **Complete workforce management platform built on EntityDB v2.32.4**

## 📁 Directory Structure

```
worca/
├── index.html              # 🏠 Main application entry point
├── README.md               # 📖 This file
│
├── 📂 bootstrap/           # 🌱 Data bootstrapping & initialization
│   ├── dataset-manager.js  # Multi-workspace dataset management
│   ├── sample-data.js      # Template-based sample data generation
│   └── schema-validator.js # Data validation and integrity checks
│
├── 📂 config/             # ⚙️ Configuration & API integration
│   ├── defaults.json       # Default configuration values
│   ├── entitydb-client.js  # EntityDB v2.32.4 API client wrapper
│   └── worca-config.js     # Configuration management system
│
├── 📂 css/                # 🎨 Stylesheets (if needed)
│
├── 📂 docs/               # 📚 Documentation
│   ├── README.md           # Main documentation
│   ├── README-INTEGRATION.md # Integration guide
│   ├── TESTING.md          # Testing procedures
│   └── demo.md             # Demo scenarios
│
├── 📂 js/                 # 💻 JavaScript modules
│   ├── worca-api.js        # Core business logic & EntityDB integration
│   ├── worca-events.js     # Real-time synchronization & events
│   ├── worca-widgets.js    # UI widgets & components
│   └── worca.js            # Main Alpine.js application logic
│
├── 📂 resources/          # 🎭 Assets & static resources
│   ├── worca-icon.svg      # Application favicon
│   ├── worca-logo-dark.svg # Dark theme logo
│   └── worca-logo-light.svg # Light theme logo
│
└── 📂 tools/              # 🔧 Utilities & development tools
    ├── bootstrap.html      # Sample data bootstrap interface
    ├── fix-config.html     # Configuration troubleshooting
    └── test-bootstrap.html # Bootstrap validation & testing
```

## 🚀 Quick Start

1. **Access Worca**: Navigate to `/worca/` in your EntityDB installation
2. **Login**: Use your EntityDB credentials
3. **Bootstrap Data**: Visit `/worca/tools/bootstrap.html` for sample data
4. **Start Working**: Create organizations, projects, epics, stories, and tasks!

## 🔧 Configuration

Worca automatically detects your EntityDB server configuration. For manual configuration or troubleshooting, use:
- `/worca/tools/fix-config.html` - Fix connection issues
- `/worca/tools/test-bootstrap.html` - Validate integration

## 📖 Documentation

See the `docs/` directory for comprehensive documentation:
- **Integration Guide**: How Worca integrates with EntityDB
- **Testing Guide**: Validation and troubleshooting procedures
- **Demo Scenarios**: Example use cases and workflows

## 🏗️ Architecture

**Frontend**: Alpine.js + Chart.js + Modern CSS
**Backend**: EntityDB v2.32.4 Temporal Database
**Integration**: Real-time synchronization with 5-second polling
**Authentication**: JWT Bearer tokens with automatic refresh
**Data Model**: Tag-based entities with binary content storage

## 🎯 Features

- ✅ **Complete Workforce Management**: Organizations → Projects → Epics → Stories → Tasks
- ✅ **Real-time Collaboration**: Live updates with conflict resolution
- ✅ **Temporal Database**: Full history tracking with nanosecond precision
- ✅ **RBAC Integration**: EntityDB role-based access control
- ✅ **Multi-workspace Support**: Dataset isolation and management
- ✅ **Professional UI**: Dark/light themes, responsive design
- ✅ **Sample Data Bootstrap**: Quick setup with realistic examples

## 🔗 Related

- **EntityDB**: https://git.home.arpa/itdlabs/entitydb.git
- **Version**: v2.32.4 integration
- **License**: Same as EntityDB project

---

*Built with ❤️ on EntityDB temporal database platform*