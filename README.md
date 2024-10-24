# CyberDitto: Digital Twin Security Benchmarking and Emulation

## Overview
CyberDitto provides an automated platform for creating digital twins of real-world systems, executing CIS benchmarks on Windows endpoints, and running security emulation scenarios using Caldera. The solution helps identify security gaps, generates remediation recommendations, and offers comprehensive reporting for users to enhance their security posture.

## Table of Contents
- [Features](#features)
- [Technologies Used](#technologies-used)
- [Architecture](#architecture)
- [Setup Instructions](#setup-instructions)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Features
1. **User Registration & Onboarding**
   - Secure user authentication via Supabase (email/password, OAuth).
   - Guided onboarding process for understanding CIS benchmarks and Caldera simulations.

2. **Digital Twin Creation**
   - Create a digital twin of the network architecture using NetBox.
   - Integrate data from Ansible for real device configurations.

3. **CIS Benchmark Automation**
   - Execute CIS benchmarks on Windows endpoints using PowerShell scripts.
   - Identify compliance gaps based on benchmark recommendations.

4. **Caldera Emulation Execution**
   - Run Caldera attack simulations based on identified security gaps or predefined attack scenarios.
   - Monitor and log attack results for analysis.

5. **Data Analysis & Reporting**
   - Comprehensive reporting on CIS benchmark results, gap analysis, and Caldera emulation outcomes.
   - Real-time dashboards for monitoring security posture and vulnerabilities.

6. **Remediation and Automated Fixes**
   - Generate remediation recommendations and scripts/commands for fixing compliance issues.
   - One-click execution of remediation actions.

## Technologies Used
- **Frontend**: React.js, Vite, TypeScript
- **Backend**: Python
- **Database**: Supabase (PostgreSQL)
- **Automation**: PowerShell for CIS Benchmark execution
- **Security Emulation**: Caldera
- **Containerization**: Docker

## Architecture
The project consists of the following components:
1. **Dashboard**: Display user architecture, compliance status, and emulation results.
2. **Digital Twin**: Visual representation of the user’s network setup.
3. **Benchmark Module**: Automate CIS Benchmark checks and gap analysis.
4. **Emulation Module**: Execute Caldera attack simulations based on gaps.
5. **Report Module**: View past benchmarks and emulation reports.

## Setup Instructions
1. Clone the repository:
   ```bash
   git clone https://github.com/sshreyx1/CyberDitto.git
   cd CyberDitto