# Nook Service Roadmap

This file tracks the current development priorities and future plans for the Nook metadata service. Update this file as tasks are completed or new priorities emerge.

## Completed (2025)
- [x] Implement basic machine CRUD operations
- [x] Add SSH key management endpoints
- [x] Add network management endpoints (placeholder)
- [x] Implement cascade deletion for SSH keys
- [x] Add CLI commands for resource management
- [x] Fix all errcheck linting issues
- [x] Establish CI pipeline with testing and coverage
- [x] Test machine CRUD operations via API
- [x] Document architecture and design decisions
- [x] Implement repository pattern with clean interfaces

## Current Status (September 2025)
- ✅ **Server runs successfully** on port 8080
- ✅ **Machine CRUD operations tested** and working
- ✅ **Database migrations** and schema management working
- ✅ **Cascade deletion** implemented and tested
- ✅ **CLI framework** established with subcommands
- ✅ **Error handling** consistent across all endpoints
- ✅ **Documentation** updated to reflect current state

## Next Steps (Immediate Priority)
- [ ] Complete SSH key endpoint testing
- [ ] Complete network endpoint testing
- [ ] Test cloud-init metadata endpoints
- [ ] Test CLI commands end-to-end
- [ ] Verify cascade deletion behavior
- [ ] Add integration tests for CLI operations

## Future Enhancements
- [ ] Improve error handling and validation
- [ ] Add more integration tests for edge/error cases
- [ ] Add security and access control (tokens, logging, etc.)
- [ ] Update documentation and developer experience
- [ ] Add deployment automation (Ansible, CI/CD)
- [ ] Expand metadata endpoints (revisit after real-world testing)
- [ ] Ensure all new features and changes follow the validation workflow (see README.md)
- [ ] Require documentation and tests for all new endpoints/features

---

**Development Philosophy:**
- Focus on working, tested features over perfect coverage metrics
- Test real-world scenarios with actual HTTP requests
- Document architecture decisions and trade-offs
- Maintain clean, idiomatic Go code following established patterns
- Prioritize reliability and maintainability over feature velocity

**How to use:**
- Check off items as you complete them.
- Add notes, links, or details for each task as needed.
- Revisit and reprioritize after real-world testing or feedback.
