# KWeaver

[中文](README.zh.md) | English

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt)

KWeaver is an open-source ecosystem for building, deploying, and running decision intelligence AI applications. This ecosystem adopts ontology as the core methodology for business knowledge networks, with DIP as the core platform, aiming to provide elastic, agile, and reliable enterprise-grade decision intelligence to further unleash everyone's productivity.

The DIP platform includes key subsystems such as ADP, Decision Agent, DIP Studio, and AI Store.

## 📚 Quick Links

- 🤝 [Contributing](rules/CONTRIBUTING.md) - Guidelines for contributing to the project
- 🚀 [Release Guidelines](rules/RELEASE.md) - Version management and release process
- 📄 [License](LICENSE.txt) - Apache License 2.0
- 🐛 [Report Bug](https://github.com/kweaver-ai/kweaver/issues) - Report a bug or issue
- 💡 [Request Feature](https://github.com/kweaver-ai/kweaver/issues) - Suggest a new feature

## Platform Architecture

```text
┌─────────────────────────────────────────────┐
│              DIP Platform                   │
│  ┌───────────────────────────────────────┐  │
│  │             AI Store                  │  │
│  ├───────────────────────────────────────┤  │
│  │            DIP Studio                 │  │
│  ├───────────────────────────────────────┤  │
│  │          Decision Agent               │  │
│  ├───────────────────────────────────────┤  │
│  │               ADP                     │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### Core Subsystems

| Sub-project | Description | Repository |
| --- | --- | --- |
| **AI Store** | AI application and component marketplace | [kweaver-ai/ai-store](https://github.com/kweaver-ai/ai-store) |
| **Studio** | DIP Studio - Visual development and management interface | [kweaver-ai/studio](https://github.com/kweaver-ai/studio) |
| **Decision Agent** | Intelligent decision agent | [kweaver-ai/decision-agent](https://github.com/kweaver-ai/decision-agent) |
| **ADP** | AI Data Platform - Core development framework, including Ontology Engine, ContextLoader, and VEGA data virtualization engine | [kweaver-ai/adp](https://github.com/kweaver-ai/adp) |
| **Operator Hub** | Operator management and orchestration platform | [kweaver-ai/operator-hub](https://github.com/kweaver-ai/operator-hub) |
| **Sandbox** | Sandbox runtime environment | [kweaver-ai/sandbox](https://github.com/kweaver-ai/sandbox) |

## Contributing

We welcome contributions! Please see our [Contributing Guide](rules/CONTRIBUTING.md) for details on how to contribute to this project.

Quick start:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE.txt) file for details.

## Support & Contact

- **Contributing**: [Contributing Guide](rules/CONTRIBUTING.md)
- **Issues**: [GitHub Issues](https://github.com/kweaver-ai/kweaver/issues)
- **License**: [Apache License 2.0](LICENSE.txt)

---

More components will be open-sourced in the future. Stay tuned!
