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
- [x] Add server subcommand with configurable database and port
- [x] Implement systemd user service for production deployment
- [x] Enhance integration testing with automatic cleanup
- [x] Streamline API to focus on nocloud compatibility
- [x] Remove unused EC2-style endpoints (public-keys, instance-identity)
- [x] Remove unused networks endpoints and handlers
- [x] Complete SSH key endpoint testing
- [x] Complete cloud-init metadata endpoint testing
- [x] Add comprehensive error handling and validation
- [x] Implement network-based IP allocation system
- [x] Add IP lease management with conflict detection
- [x] Create VM provisioning automation scripts
- [x] Integrate SSH key management with cloud-init user-data
- [x] Add database migrations for network and IP lease tables
- [x] Test IP allocation on real homelab virt bridge network
- [x] Create comprehensive VM provisioning documentation

## Current Status (September 2025)
- ✅ **Production deployment ready** with systemd user service
- ✅ **Server runs successfully** on port 8080 with configurable options
- ✅ **Machine CRUD operations tested** and working
- ✅ **SSH key management fully tested** with comprehensive coverage
- ✅ **Cloud-init metadata endpoints tested** with IP-based lookup
- ✅ **Database migrations** and schema management working
- ✅ **Cascade deletion** implemented and tested
- ✅ **CLI framework** established with server subcommand
- ✅ **Error handling** consistent across all endpoints
- ✅ **Integration testing** with automatic server lifecycle management
- ✅ **Documentation** updated to reflect current streamlined architecture
- ✅ **Network-based IP allocation** implemented and tested on homelab
- ✅ **IP lease management** with conflict detection working
- ✅ **VM provisioning automation** scripts created and tested
- ✅ **SSH key integration** with cloud-init user-data verified
- ✅ **Comprehensive documentation** for VM provisioning workflow

## Next Steps (Immediate Priority)
- [ ] Test CLI commands end-to-end
- [ ] Verify cascade deletion behavior in integration tests
- [ ] Add integration tests for CLI operations
- [ ] Add security and access control (tokens, logging, etc.)
- [ ] Improve error handling and validation for edge cases

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
