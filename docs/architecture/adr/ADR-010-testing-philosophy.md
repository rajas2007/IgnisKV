# ADR-010: Testing Philosophy

## Status
Accepted

## Date
2026-07-02

## Context
Infrastructure projects like databases require high stability and reliability. A minor bug in the parser or store lock coordination can cause severe data corruption, deadlocks, or security issues. Simply writing tests after development or relying solely on end-to-end integration tests makes it difficult to locate bugs and verify component boundaries.

## Problem Statement
What methodology and structure should be adopted for testing IgnisKV to guarantee engine correctness and prevent regression bugs?

## Decision
We establish a structured development and testing cycle:
1. **Developer Workflow**:
   
   `Design -> Implement -> Manual Verification -> Unit Tests -> Commit`
   
   Every commit containing logic modifications must include accompanying unit tests.
2. **Component Test Isolation**: We test core components independently to ensure that tests focus purely on the target logic:
   * **Store**: Verified independently for keyspace operations, lock safety, and boundary checks.
   * **Parser**: Tokenization tests verify string splitting, argument slicing, quote handling, and invalid syntax errors without loading a store.
   * **Dispatcher**: Tests verify command registration, lookup routing, and missing handler errors.
   * **Command Handlers**: Handlers are tested by mocking or using a localized concrete store, checking expected responses for valid/invalid parameters.
   * **Network (Server)**: Socket connection handshakes, protocol timeouts, and connection limit tests run in separate integration tests under `internal/server`.

## Rationale
* Testing in isolation ensures that when a test fails, the bug is immediately located inside the failing component, rather than having to trace through a full client-server stack.
* Establishing the development loop as a rule helps prevent the accumulation of untested features.

## Alternatives Considered
* **Alternative A: Rely on end-to-end integration tests**: Spin up the TCP server, execute commands via a local CLI client, and check stdout. Rejected because integration tests are slow, brittle, and do not test edge case errors inside internal components cleanly.
* **Alternative B: Write tests only at the end of the project**: Rejected because bugs are significantly harder and more expensive to resolve once multiple layers have been built on top of them.

## Consequences
* **Positive**: 100% confidence in component correctness, fast test suite execution, and easy local debugging.
* **Negative**: Writing isolated tests requires mock structures and test setup boilerplate for each package.

## Future Evolution
In future milestones, we will integrate concurrency load test suites and continuous integration (CI) workflows to run tests automatically on every PR.
