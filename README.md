<div id="top">

<!-- HEADER STYLE: CLASSIC -->
<div align="center">

<img src=".github/images/icon.jpeg" width="30%" style="position: relative; top: 0; right: 0;" alt="Project Logo"/>

# <code>❯ REPLACE-ME</code>

<em>Edge telemetry, reliably delivered at scale</em>

<!-- BADGES -->
<!-- local repository, no metadata badges. -->

<em>Built with the tools and technologies:</em>

<img src="https://img.shields.io/badge/Zsh-F15A24.svg?style=default&logo=Zsh&logoColor=white" alt="Zsh">
<img src="https://img.shields.io/badge/Go-00ADD8.svg?style=default&logo=Go&logoColor=white" alt="Go">
<img src="https://img.shields.io/badge/YAML-CB171E.svg?style=default&logo=YAML&logoColor=white" alt="YAML">

</div>
<br>

---

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Overview](#overview)
- [Features](#features)
- [Project Structure](#project-structure)
    - [Project Index](#project-index)
- [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Usage](#usage)
    - [Testing](#testing)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## Overview

**Carrier Pigeons** is a lightweight, production-ready telemetry pipeline that collects system metrics from edge devices and reliably delivers them to centralized storage via MQTT—complete with robust state management, batch processing, and graceful lifecycle handling.

**Why Carrier Pigeons?**

This project empowers developers to build scalable, observable infrastructure monitoring systems without reinventing core pub/sub, storage, and resilience patterns. The core features include:

- **🔋 Cross-platform metric collection:** Gathers CPU temperature/frequency (Linux) and memory metrics with platform-aware fallbacks for macOS
- **📡 MQTT pub/sub architecture:** Uses Eclipse Paho for reliable, at-least-once delivery with automatic reconnection and exponential backoff
- **💾 PostgreSQL storage layer:** Batch-inserts telemetry with connection pooling, schema auto-initialization, and health checks
- **⚙️ Configurable worker pools:** Parallel processing with tunable batch size, concurrency, and flush intervals for high-throughput ingestion
- **🔄 State machine orchestration:** Thread-safe connection lifecycle management with readiness signals and transition callbacks
- **📝 YAML-driven configuration:** Centralized, validated settings for MQTT, storage, workers, and metrics exposure

---

## Features

|      | Component       | Details                                                                                                                                                                                                 |
| :--- | :-------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ⚙️  | **Architecture**  | <ul><li>Multi-module Go monorepo: contains at least two modules: `carrierpigeons/` and `pigeoncoop/`, each with their own <code>go.mod</code></li><li>Modular service design implied by module separation (e.g., core vs. internal components)</li><li>Uses standard Go module layout: <code>cmd/</code>, <code>internal/</code>, <code>pkg/</code> likely present</li></ul> |
| 🔩 | **Code Quality**  | <ul><li>Relies on Go’s built-in tooling: <code>go vet</code>, <code>gofmt</code>, and <code>goimports</code> (implied by <code>go.mod</code> usage)</li><li>Linter support likely via <code>golangci-lint</code> or similar (not explicit, but standard for Go projects)</li></ul> |
| 📄 | **Documentation** | <ul><li>Includes <code>config.example.yaml</code> — suggests runtime configuration via YAML</li><li>Lacks explicit docs folder or README in provided context; may rely on inline comments and example config</li><li><code>readmeai-prompt.zsh</code> hints at automated README generation (possibly via README.ai)</li></ul> |
| 🔌 | **Integrations**  | <ul><li>Uses <code>paho.golang</code> → indicates integration with <strong>Eclipse Paho MQTT client</strong> for IoT/messaging</li><li>Uses <code>pgpassfile</code>, <code>pgservicefile</code> → implies PostgreSQL connectivity via libpq (e.g., <code>lib/pq</code> or <code>pgx</code>)</li><li><code>crypto</code> and <code>sync</code> from stdlib used for secure operations and concurrency</li></ul> |
| 🧩 | **Modularity**    | <ul><li>Explicit Go modules: <code>carrierpigeons/go.mod</code>, <code>pigeoncoop/go.mod</code></li><li>Internal packages likely separated under <code>internal/</code> to enforce encapsulation</li><li>Config-driven architecture via <code>config.example.yaml</code> enables runtime modularity</li></ul> |
| 🧪 | **Testing**       | <ul><li>No explicit test files or <code>_test.go</code> mentioned in context</li><li>Standard Go testing framework available (<code>go test</code>)</li><li>Likely uses table-driven tests (Go best practice)</li></ul> |

---

## Project Structure

```sh
└── /
    ├── .github
    │   └── images
    ├── LICENSE
    ├── README.md
    ├── carrierpigeons
    │   ├── README.md
    │   ├── carrierpigeons
    │   ├── cmd
    │   ├── config.example.yaml
    │   ├── go.mod
    │   ├── go.sum
    │   ├── internal
    │   └── pkg
    ├── pigeoncoop
    │   ├── README.md
    │   ├── cmd
    │   ├── config.example.yaml
    │   ├── go.mod
    │   ├── go.sum
    │   ├── internal
    │   └── pkg
    ├── readme-ai.md
    └── readmeai-prompt.zsh
```

### Project Index

<details open>
	<summary><b><code>/</code></b></summary>
	<!-- __root__ Submodule -->
	<details>
		<summary><b>__root__</b></summary>
		<blockquote>
			<div class='directory-path' style='padding: 8px 0; color: #666;'>
				<code><b>⦿ __root__</b></code>
			<table style='width: 100%; border-collapse: collapse;'>
			<thead>
				<tr style='background-color: #f8f9fa;'>
					<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
					<th style='text-align: left; padding: 8px;'>Summary</th>
				</tr>
			</thead>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/LICENSE'>LICENSE</a></b></td>
					<td style='padding: 8px;'>- The MIT License establishes the terms under which the software may be used, modified, and distributed by anyone, ensuring broad accessibility and permissive reuse while disclaiming warranties and limiting liability<br>- It enables contributors and users to freely integrate the project into both open-source and proprietary systems, fostering collaboration and innovation without legal restrictions—provided the original copyright notice is retained.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/readmeai-prompt.zsh'>readmeai-prompt.zsh</a></b></td>
					<td style='padding: 8px;'>- Generates project documentation by analyzing the local repository using the Ollama API with the godocu-qwen model, while incorporating a custom logo for branding<br>- It streamlines README creation for developers seeking consistent, AI-powered documentation without manual effort or external dependencies beyond the specified configuration.</td>
				</tr>
			</table>
		</blockquote>
	</details>
	<!-- carrierpigeons Submodule -->
	<details>
		<summary><b>carrierpigeons</b></summary>
		<blockquote>
			<div class='directory-path' style='padding: 8px 0; color: #666;'>
				<code><b>⦿ carrierpigeons</b></code>
			<table style='width: 100%; border-collapse: collapse;'>
			<thead>
				<tr style='background-color: #f8f9fa;'>
					<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
					<th style='text-align: left; padding: 8px;'>Summary</th>
				</tr>
			</thead>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/carrierpigeons/go.mod'>go.mod</a></b></td>
					<td style='padding: 8px;'>- CarrierPigeons defines the Go module for a messaging service that routes messages via MQTT using the Eclipse Paho client and manages configuration through YAML files<br>- It serves as the foundational dependency declaration for the carrierpigeons subpackage, enabling reliable, asynchronous communication within the broader system architecture while ensuring version consistency across builds.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/carrierpigeons/carrierpigeons'>carrierpigeons</a></b></td>
					<td style='padding: 8px;'>Carrie<code> are incomplete or malformed, making it impossible to determine the actual file or project structure.To deliver an accurate, succinct summary of the file’s purpose within the architecture, I’ll need:-The <strong>full project structure</strong> (e.g., output of </code>tree<code> or a list of key directories/files), and-Clarification: Is </code>carrie<code> the filename (e.g., </code>carrie.py<code>, </code>carrie.js`)? Or is it a directory?Once you provide those, I’ll give you a high-level, architecture-focused summary—what the file <em>does</em> in service of the system, not <em>how</em> it does it.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/carrierpigeons/go.sum'>go.sum</a></b></td>
					<td style='padding: 8px;'>- This <code>go.sum</code> file ensures reproducible builds by locking exact versions and cryptographic hashes of all Go dependencies used across the project, including core libraries like Eclipse Paho for MQTT communication, Google’s test utilities, and YAML parsing tools<br>- It guarantees consistency across environments by verifying dependency integrity during builds, forming a foundational layer for reliable and predictable software delivery within the broader codebase architecture.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/carrierpigeons/config.example.yaml'>config.example.yaml</a></b></td>
					<td style='padding: 8px;'>- CarrierPigeons configuration defines the identity and operational parameters for an edge telemetry node, specifying its unique identifier, enabled metric collectors, and connection details for publishing data to a message broker<br>- It also configures internal buffering capacities to ensure smooth metric collection and outbound transmission, aligning with a distributed, real-time monitoring architecture focused on lightweight, reliable data flow from edge devices to centralized systems.</td>
				</tr>
			</table>
			<!-- cmd Submodule -->
			<details>
				<summary><b>cmd</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ carrierpigeons.cmd</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='/carrierpigeons/cmd/main.go'>main.go</a></b></td>
							<td style='padding: 8px;'>- CarrierPigeons daemon initializes and orchestrates a telemetry collection and publishing pipeline, loading configuration, spinning up metric collectors (CPU and memory), establishing an MQTT connection, and starting a scheduler to manage data flow<br>- It ensures graceful startup and shutdown with state transitions and robust error handling, serving as the central control plane for the system’s real-time telemetry ingestion and delivery.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- internal Submodule -->
			<details>
				<summary><b>internal</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ carrierpigeons.internal</b></code>
					<!-- publisher Submodule -->
					<details>
						<summary><b>publisher</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ carrierpigeons.internal.publisher</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/publisher/mqtt_test.go'>mqtt_test.go</a></b></td>
									<td style='padding: 8px;'>Validates the MQTT publisher’s initialization logic and state behavior within the Carrier Pigeons system, ensuring correct configuration handling—including defaults—and proper transition to the expected initial disconnected state via the shared state machine.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/publisher/mqtt.go'>mqtt.go</a></b></td>
									<td style='padding: 8px;'>The MQTT publisher implements reliable telemetry delivery by establishing persistent connections to an MQTT broker with automatic reconnection and exponential backoff, ensuring metrics are published with at-least-once delivery semantics while maintaining connection state awareness and graceful shutdown behavior across the system’s pub-sub architecture.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/publisher/interface.go'>interface.go</a></b></td>
									<td style='padding: 8px;'>- The interface defines the contract for publishing telemetry metrics within the carrierpigeons system, enabling decoupled integration with various telemetry backends<br>- It specifies core lifecycle operations—starting, stopping, and publishing metrics—alongside connection state checks, ensuring reliable and graceful data transmission<br>- The included Metric and configuration types standardize data structure and setup, supporting consistent metric handling across the observability pipeline.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- collector Submodule -->
					<details>
						<summary><b>collector</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ carrierpigeons.internal.collector</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/collector/cpu_darwin.go'>cpu_darwin.go</a></b></td>
									<td style='padding: 8px;'>A macOS-specific CPU metric collector that integrates into the broader system monitoring architecture by fulfilling the collector interface contract while gracefully handling platform limitations—specifically acknowledging that CPU temperature data is unavailable on Darwin systems, thus enabling cross-platform compatibility without breaking the build.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/collector/cpu.go'>cpu.go</a></b></td>
									<td style='padding: 8px;'>- The CPUCollector gathers Linux-specific CPU temperature and frequency metrics by reading thermal zones and cpufreq interfaces, enriching each metric with configurable labels and zone-specific metadata<br>- It operates on a 10-second interval, gracefully handles missing hardware paths, and prioritizes context-aware cancellation during collection to ensure responsive, reliable system monitoring within the broader carrier pigeon telemetry pipeline.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/collector/interface.go'>interface.go</a></b></td>
									<td style='padding: 8px;'>- Defines the core abstraction for system metric collection across the carrierpigeons architecture, enabling pluggable collectors to gather standardized metrics like CPU usage, memory, disk, and service states<br>- It establishes a consistent interface for retrieving time-stamped metric data with node context and metadata, supporting flexible configuration and context-aware collection intervals.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/collector/memory_linux.go'>memory_linux.go</a></b></td>
									<td style='padding: 8px;'>- Collects Linux memory utilization metrics by parsing <code>/proc/meminfo</code>, computing usage and availability percentages, and emitting structured metrics with configurable labels<br>- Designed for periodic collection every 15 seconds, it efficiently handles context cancellation and minimizes garbage collection pressure through buffer reuse<br>- Integrates into the broader collector framework to feed system health data into the carrierpigeon monitoring pipeline.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/collector/memory_darwin.go'>memory_darwin.go</a></b></td>
									<td style='padding: 8px;'>A macOS-specific placeholder implementation that satisfies the collector interface without actually gathering memory metrics, ensuring cross-platform compatibility by allowing the application to compile on Darwin while deliberately returning an error to signal unimplemented functionality.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- scheduler Submodule -->
					<details>
						<summary><b>scheduler</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ carrierpigeons.internal.scheduler</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/scheduler/scheduler.go'>scheduler.go</a></b></td>
									<td style='padding: 8px;'>- The scheduler orchestrates periodic metric collection from diverse sources and publishes them in batches to external systems, managing backpressure through buffered channels and graceful shutdown via context cancellation<br>- It coordinates concurrent collectors, aggregates metrics into configurable batches, and tracks operational metrics like dropped or published counts for observability.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- state Submodule -->
					<details>
						<summary><b>state</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ carrierpigeons.internal.state</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/state/machine_test.go'>machine_test.go</a></b></td>
									<td style='padding: 8px;'>Verifies the correctness and thread safety of a finite state machine managing carrier pigeon connection states, ensuring reliable transitions between Disconnected, Connecting, Connected, and Backoff phases while accurately tracking backoff durations, readiness signals, and connection metrics under concurrent access.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/state/machine.go'>machine.go</a></b></td>
									<td style='padding: 8px;'>- The state machine manages MQTT connection lifecycle transitions—disconnected, connecting, connected, reconnecting, and backoff—with built-in exponential backoff, jitter, and metrics tracking<br>- It enforces valid state transitions, triggers callbacks on changes, and exposes readiness and stop signals for coordination with other components in the carrierpigeons system.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/carrierpigeons/internal/state/machine_benchmark_test.go'>machine_benchmark_test.go</a></b></td>
									<td style='padding: 8px;'>- Performance benchmarks validate the state machine’s core operations under realistic workloads, measuring transition speed, backoff duration calculations, and jittered retry behavior<br>- Concurrent benchmarks assess thread safety and allocation efficiency during high-frequency state changes across multiple goroutines<br>- These tests ensure the state machine remains responsive and reliable under load, supporting robust connection management in the carrierpigeons system.</td>
								</tr>
							</table>
						</blockquote>
					</details>
				</blockquote>
			</details>
		</blockquote>
	</details>
	<!-- pigeoncoop Submodule -->
	<details>
		<summary><b>pigeoncoop</b></summary>
		<blockquote>
			<div class='directory-path' style='padding: 8px 0; color: #666;'>
				<code><b>⦿ pigeoncoop</b></code>
			<table style='width: 100%; border-collapse: collapse;'>
			<thead>
				<tr style='background-color: #f8f9fa;'>
					<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
					<th style='text-align: left; padding: 8px;'>Summary</th>
				</tr>
			</thead>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/pigeoncoop/go.mod'>go.mod</a></b></td>
					<td style='padding: 8px;'>- This Go module defines the core dependency manifest for the pigeoncoop package, establishing it as a component within the Carrier-Pigeons project<br>- It declares direct dependencies on MQTT client libraries, PostgreSQL database drivers, and YAML parsing utilities—indicating its role in handling message brokering, data persistence, and configuration management across the system’s infrastructure.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/pigeoncoop/go.sum'>go.sum</a></b></td>
					<td style='padding: 8px;'>- This <code>go.sum</code> file ensures reproducible builds by locking exact versions and cryptographic checksums of all direct and transitive dependencies used across the Go codebase, including core libraries like <code>paho.golang</code> for MQTT communication, <code>pgx/v5</code> for PostgreSQL access, and testing utilities such as <code>testify</code><br>- It acts as a security and consistency checkpoint, verifying that dependency artifacts match expected versions during builds.</td>
				</tr>
				<tr style='border-bottom: 1px solid #eee;'>
					<td style='padding: 8px;'><b><a href='/pigeoncoop/config.example.yaml'>config.example.yaml</a></b></td>
					<td style='padding: 8px;'>- PigeonCoop configuration defines core system parameters for MQTT connectivity, database storage, worker processing, and metrics exposure<br>- It establishes how the application ingests telemetry data via MQTT, stores it in PostgreSQL, processes messages with parallel workers, and exposes Prometheus-compatible metrics—ensuring reliable, scalable, and observable operation across the entire pipeline.</td>
				</tr>
			</table>
			<!-- cmd Submodule -->
			<details>
				<summary><b>cmd</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ pigeoncoop.cmd</b></code>
					<table style='width: 100%; border-collapse: collapse;'>
					<thead>
						<tr style='background-color: #f8f9fa;'>
							<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
							<th style='text-align: left; padding: 8px;'>Summary</th>
						</tr>
					</thead>
						<tr style='border-bottom: 1px solid #eee;'>
							<td style='padding: 8px;'><b><a href='/pigeoncoop/cmd/main.go'>main.go</a></b></td>
							<td style='padding: 8px;'>Initializes and launches the PigeonCoop application by loading configuration from a YAML file, instantiating the core application logic, and managing its lifecycle—including graceful shutdown—ensuring robust startup, error handling, and clean termination across the entire system.</td>
						</tr>
					</table>
				</blockquote>
			</details>
			<!-- internal Submodule -->
			<details>
				<summary><b>internal</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ pigeoncoop.internal</b></code>
					<!-- config Submodule -->
					<details>
						<summary><b>config</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ pigeoncoop.internal.config</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/config/config.go'>config.go</a></b></td>
									<td style='padding: 8px;'>- Configuration loading and validation for the PigeonCoop application, centralizing settings for MQTT connectivity, database access, worker pool behavior, and metrics collection from a YAML file<br>- It ensures required values are present, applies sensible defaults where missing, and provides typed conversion helpers to feed downstream components like pubsub, storage, and worker modules—forming the foundational configuration layer that drives the entire system’s runtime behavior.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/config/config_test.go'>config_test.go</a></b></td>
									<td style='padding: 8px;'>- Validates configuration loading and structure for the pigeoncoop system, ensuring YAML-based settings for MQTT, storage, workers, and metrics are correctly parsed and validated<br>- It verifies required fields like broker URL and database host, confirms config transformation to client-specific formats, and safeguards against invalid or incomplete configurations before runtime.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- state Submodule -->
					<details>
						<summary><b>state</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ pigeoncoop.internal.state</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/state/state_test.go'>state_test.go</a></b></td>
									<td style='padding: 8px;'>- Validates the state machine’s transition rules, backoff behavior, and readiness semantics to ensure reliable connection lifecycle management across the pigeoncoop system<br>- It confirms transitions like disconnected → connecting → connected are permitted while blocking invalid ones like connected → connecting, verifies exponential backoff resets correctly, and checks that the Ready channel closes only upon reaching the Connected state.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/state/machine.go'>machine.go</a></b></td>
									<td style='padding: 8px;'>- The state machine manages connection lifecycle transitions for MQTT and database clients, enforcing valid state changes like Disconnected → Connecting → Connected and handling exponential backoff during failures<br>- It exposes readiness and stop signals via channels, supports transition callbacks, and ensures thread-safe state operations with configurable retry behavior—enabling robust, predictable connection management across the system.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- storage Submodule -->
					<details>
						<summary><b>storage</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ pigeoncoop.internal.storage</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/storage/interface.go'>interface.go</a></b></td>
									<td style='padding: 8px;'>- Defines the storage interface for telemetry data persistence, enabling decoupled database interactions within the pigeoncoop architecture<br>- It specifies core operations—initialization, batch and single record insertion, graceful shutdown, and health checks—while abstracting PostgreSQL-specific logic<br>- This interface supports consistent telemetry ingestion and reliability across services by standardizing how telemetry records are stored and verified.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/storage/postgres.go'>postgres.go</a></b></td>
									<td style='padding: 8px;'>- PostgresStore implements a robust PostgreSQL-backed telemetry storage layer, enabling efficient batch and single record ingestion with automatic schema initialization and connection pooling<br>- It ensures reliable health checks, graceful shutdowns, and scalable data persistence for distributed telemetry systems by leveraging pgx’s high-performance features and transaction-safe operations.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- pubsub Submodule -->
					<details>
						<summary><b>pubsub</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ pigeoncoop.internal.pubsub</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/pubsub/mqtt.go'>mqtt.go</a></b></td>
									<td style='padding: 8px;'>An MQTT client implementation enabling asynchronous message publishing and subscription via the Eclipse Paho Go library, bridging application components with external MQTT brokers for real-time event-driven communication within the pubsub subsystem.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/pubsub/interface.go'>interface.go</a></b></td>
									<td style='padding: 8px;'>- The interface defines the contract for MQTT-based message broker interactions within the pubsub package, enabling decoupled communication across services<br>- It standardizes connection management, topic subscription, and message publishing—ensuring consistent behavior regardless of the underlying MQTT implementation<br>- This abstraction supports reliable, asynchronous messaging as a core communication layer in the system’s architecture.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- telemetry Submodule -->
					<details>
						<summary><b>telemetry</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ pigeoncoop.internal.telemetry</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/telemetry/telemetry_test.go'>telemetry_test.go</a></b></td>
									<td style='padding: 8px;'>- Validates telemetry payloads for correctness and ensures reliable JSON serialization/deserialization across the system<br>- It safeguards data integrity by enforcing required fields like node ID and metric type, while confirming structured data remains consistent through encoding and decoding cycles—critical for robust telemetry ingestion and processing within the broader distributed monitoring architecture.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/telemetry/types.go'>types.go</a></b></td>
									<td style='padding: 8px;'>- TelemetryPayload defines the standardized structure for telemetry events collected from edge nodes, ensuring consistent data flow across the PigeonCoop system<br>- It enforces validation of critical fields like node identification and metric type to maintain data integrity before further processing or storage.</td>
								</tr>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/telemetry/json.go'>json.go</a></b></td>
									<td style='padding: 8px;'>- Custom JSON serialization for telemetry payloads ensures consistent timestamp formatting using RFC3339Nano and seamless round-trip encoding/decoding across the system<br>- It standardizes how time-sensitive telemetry data is serialized to JSON and deserialized back into structured Go types, enabling reliable interoperability with external systems and storage layers while preserving precision and timezone-awareness.</td>
								</tr>
							</table>
						</blockquote>
					</details>
					<!-- worker Submodule -->
					<details>
						<summary><b>worker</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ pigeoncoop.internal.worker</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/internal/worker/pool.go'>pool.go</a></b></td>
									<td style='padding: 8px;'>- A worker pool that efficiently batches and inserts telemetry records into storage using configurable concurrency, batch size, and flush intervals<br>- It accepts records via a channel, processes them in background workers, and ensures timely or full-batch writes while supporting graceful shutdown.</td>
								</tr>
							</table>
						</blockquote>
					</details>
				</blockquote>
			</details>
			<!-- pkg Submodule -->
			<details>
				<summary><b>pkg</b></summary>
				<blockquote>
					<div class='directory-path' style='padding: 8px 0; color: #666;'>
						<code><b>⦿ pigeoncoop.pkg</b></code>
					<!-- app Submodule -->
					<details>
						<summary><b>app</b></summary>
						<blockquote>
							<div class='directory-path' style='padding: 8px 0; color: #666;'>
								<code><b>⦿ pigeoncoop.pkg.app</b></code>
							<table style='width: 100%; border-collapse: collapse;'>
							<thead>
								<tr style='background-color: #f8f9fa;'>
									<th style='width: 30%; text-align: left; padding: 8px;'>File Name</th>
									<th style='text-align: left; padding: 8px;'>Summary</th>
								</tr>
							</thead>
								<tr style='border-bottom: 1px solid #eee;'>
									<td style='padding: 8px;'><b><a href='/pigeoncoop/pkg/app/app.go'>app.go</a></b></td>
									<td style='padding: 8px;'>- The application orchestrator initializes and manages core subsystems—telemetry storage, MQTT pub/sub, and worker pools—to receive, validate, and process telemetry data from connected devices<br>- It handles graceful startup and shutdown, subscribes to relevant MQTT topics, and routes incoming payloads to workers for asynchronous processing while maintaining state transitions and robust error handling throughout its lifecycle.</td>
								</tr>
							</table>
						</blockquote>
					</details>
				</blockquote>
			</details>
		</blockquote>
	</details>
</details>

---

## Getting Started

### Prerequisites

This project requires the following dependencies:

- **Programming Language:** Go
- **Package Manager:** Go modules

### Installation

Build  from the source and intsall dependencies:

1. **Clone the repository:**

    ```sh
    ❯ git clone ../
    ```

2. **Navigate to the project directory:**

    ```sh
    ❯ cd 
    ```

3. **Install the dependencies:**

<!-- SHIELDS BADGE CURRENTLY DISABLED -->
	<!-- [![go modules][go modules-shield]][go modules-link] -->
	<!-- REFERENCE LINKS -->
	<!-- [go modules-shield]: https://img.shields.io/badge/Go-00ADD8.svg?style={badge_style}&logo=go&logoColor=white -->
	<!-- [go modules-link]: https://golang.org/ -->

	**Using [go modules](https://golang.org/):**

	```sh
	❯ go build
	```

### Usage

Run the project with:

**Using [go modules](https://golang.org/):**
```sh
go run {entrypoint}
```

### Testing

 uses the {__test_framework__} test framework. Run the test suite with:

**Using [go modules](https://golang.org/):**
```sh
go test ./...
```

---

## Roadmap

- [X] **`Task 1`**: <strike>Implement feature one.</strike>
- [ ] **`Task 2`**: Implement feature two.
- [ ] **`Task 3`**: Implement feature three.

---

## Contributing

- **💬 [Join the Discussions](https://LOCAL///discussions)**: Share your insights, provide feedback, or ask questions.
- **🐛 [Report Issues](https://LOCAL///issues)**: Submit bugs found or log feature requests for the `` project.
- **💡 [Submit Pull Requests](https://LOCAL///blob/main/CONTRIBUTING.md)**: Review open PRs, and submit your own PRs.

<details closed>
<summary>Contributing Guidelines</summary>

1. **Fork the Repository**: Start by forking the project repository to your LOCAL account.
2. **Clone Locally**: Clone the forked repository to your local machine using a git client.
   ```sh
   git clone .
   ```
3. **Create a New Branch**: Always work on a new branch, giving it a descriptive name.
   ```sh
   git checkout -b new-feature-x
   ```
4. **Make Your Changes**: Develop and test your changes locally.
5. **Commit Your Changes**: Commit with a clear message describing your updates.
   ```sh
   git commit -m 'Implemented new feature x.'
   ```
6. **Push to LOCAL**: Push the changes to your forked repository.
   ```sh
   git push origin new-feature-x
   ```
7. **Submit a Pull Request**: Create a PR against the original project repository. Clearly describe the changes and their motivations.
8. **Review**: Once your PR is reviewed and approved, it will be merged into the main branch. Congratulations on your contribution!
</details>

<details closed>
<summary>Contributor Graph</summary>
<br>
<p align="left">
   <a href="https://LOCAL{///}graphs/contributors">
      <img src="https://contrib.rocks/image?repo=/">
   </a>
</p>
</details>

---

## License

 is protected under the [LICENSE](https://choosealicense.com/licenses) License. For more details, refer to the [LICENSE](https://choosealicense.com/licenses/) file.

---

## Acknowledgments

- Credit `contributors`, `inspiration`, `references`, etc.

<div align="right">

[![][back-to-top]](#top)

</div>


[back-to-top]: https://img.shields.io/badge/-BACK_TO_TOP-151515?style=flat-square


---
