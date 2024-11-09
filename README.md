# Cyber Ditto: Digital Twin Security Platform

CyberDitto is an advanced security assessment and remediation platform that creates digital twins of your network infrastructure. By leveraging tools like Nmap, Ansible, and PowerShell, it provides a safe environment to evaluate your security posture through automated CIS Benchmarking and adversarial simulations powered by Atomic Red Team. The platform automatically identifies security gaps, generates targeted remediation scripts, and provides comprehensive dashboards for monitoring your organization's security evolution - all while preserving your production environment's integrity.

## Table of Contents

1. [Features](#features)
2. [Technologies](#technologies)
3. [Installation](#installation)
4. [Usage](#usage)
   - [1. User Registration & Onboarding](#1-user-registration--onboarding)
   - [2. Digital Twin Creation](#2-digital-twin-creation)
   - [3. CIS Benchmark Automation](#3-cis-benchmark-automation)
   - [4. Attack Simulations](#4-attack-simulations)
   - [5. Data Analysis & Reporting](#5-data-analysis--reporting)
   - [6. Remediation and Automated Fixes](#6-remediation-and-automated-fixes)
5. [Configuration](#configuration)
6. [License](#license)

## Features

- **Custom User Authentication**: Secure login and registration with MySQL
- **Digital Twin Setup**: Network and endpoint replication with Nmap, Ansible, and GNS3
- **CIS Benchmarking**: Automated checks and compliance validation on Windows endpoints
- **Attack Simulation**: Real-world attack scenarios powered by Atomic Red Team
- **Comprehensive Reporting**: Real-time dashboards and detailed PDF reports
- **Automated Remediation**: Provides scripts to address detected vulnerabilities

## Technologies

- **Frontend**: React.js with Vite and TypeScript
- **Backend**: Go with MySQL, using libraries for data processing and reporting
- **Database**: MySQL
- **Network Emulation**: Nmap, Ansible, GNS3, Docker Desktop
- **Automation & Scripting**: PowerShell, Atomic Red Team
- **PDF Generation**: ReportLab or equivalent library in Go

## Installation

### Prerequisites

- [Node.js & npm](https://nodejs.org/) (for frontend)
- [Go](https://golang.org/) (for backend)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (for network emulation)
- [MySQL](https://www.mysql.com/) for authentication and data storage

### Steps

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/your-repo.git
   cd your-repo
   ```

2. **Install Frontend Dependencies**:
   ```bash
   cd frontend
   npm install
   ```

3. **Install Backend Dependencies**:
   ```bash
   # Ensure Go modules are properly set up
   cd backend
   go mod tidy
   ```

4. **Database Setup**:
   - Configure MySQL and update the backend with credentials
   - Run migrations to set up the necessary tables

5. **Start the Application**:
   ```bash
   # Launch the backend server
   go run main.go

   # Launch the frontend
   npm run dev
   ```

## Usage

### 1. User Registration & Onboarding

- **Registration/Login**: Users create accounts or log in using MySQL for secure data storage
- **Onboarding**: After login, users are guided through the onboarding process to understand digital twin creation, CIS benchmarking, and Caldera simulations

### 2. Digital Twin Creation

- **Digital Twin Setup**:
  - Users input network and endpoint details
  - Nmap scans for open ports and network details, while Ansible collects endpoint security configurations

- **Endpoint Replication**:
  - The backend orchestrates NetReplica and GNS3 to emulate the digital twin environment
  - Creates a simulated network with device configurations mirrored from the original setup

### 3. CIS Benchmark Automation

- **Benchmark Checks**:
  - PowerShell scripts run on Windows endpoints in the digital twin to perform CIS Benchmark checks
  - The backend collects configuration and compliance data for analysis

### 4. Attack Simulations

- **Scenario Selection**:
  - Users select attack scenarios from Atomic Red Team

- **Execution & Monitoring**:
  - Atomic Red Team executes the selected simulations
  - Backend logs success or failure for each step

- **Result Display**:
  - Real-time results are shown in the frontend, providing feedback on attack outcomes

### 5. Data Analysis & Reporting

- **Gap Analysis**:
  - The backend correlates compliance gaps with attack outcomes to identify high-risk vulnerabilities

- **Report Generation**:
  - A detailed report is generated, summarizing compliance results, vulnerabilities, and remediation recommendations

- **Dashboard**:
  - Real-time dashboards display compliance status, vulnerabilities, and simulation outcomes

### 6. Remediation and Automated Fixes

- **Recommendations**:
  - Based on analysis, remediation steps are provided, including PowerShell scripts to address identified issues

- **One-Click Fix Execution**:
  - Users can execute remediation scripts directly from the frontend, applying fixes to Windows endpoints

## Configuration

- **MySQL**: Configure the backend to connect to your MySQL database, storing scan data, benchmarks, and user profiles
- **Docker & GNS3**: Ensure Docker Desktop and GNS3 are set up for Windows to support network emulation

## License

This project is licensed under the MIT License.
