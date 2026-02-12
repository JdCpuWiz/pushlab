# Contributing to PushLab

Thank you for considering contributing to PushLab! 🎉

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version, iOS version)
- Relevant logs

### Suggesting Features

Feature requests are welcome! Please:
- Check existing issues first
- Describe the use case
- Explain why it would be useful
- Consider implementation complexity

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Write/update tests
5. Update documentation
6. Commit with clear messages
7. Push and create a PR

#### PR Guidelines

- Keep PRs focused on a single feature/fix
- Write clear commit messages
- Update README if needed
- Add tests for new features
- Ensure all tests pass
- Follow existing code style

### Code Style

**Go:**
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Add comments for exported functions
- Keep functions small and focused

**Swift:**
- Use Xcode formatting
- Follow [Swift API Design Guidelines](https://swift.org/documentation/api-design-guidelines/)
- Use SwiftUI best practices

### Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/pushlab
cd pushlab

# Install dependencies
make deps

# Start development environment
make dev

# Run tests
make test
```

### Commit Messages

Use conventional commits:
- `feat: Add notification templates`
- `fix: Resolve token refresh issue`
- `docs: Update API documentation`
- `test: Add device registration tests`
- `refactor: Simplify APNs client`
- `chore: Update dependencies`

### Testing

Before submitting:
- Test locally with Docker
- Verify API endpoints work
- Test iOS app on device
- Check logs for errors
- Update tests if needed

### Documentation

Update documentation for:
- New API endpoints
- Configuration changes
- Breaking changes
- New features

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Help others learn

## Questions?

- Open an issue for questions
- Check existing documentation
- Review closed issues

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for making PushLab better! 🚀
