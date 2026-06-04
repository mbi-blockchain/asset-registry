# Guide: Building and Pushing Hyperledger Fabric External Chaincode to Docker Hub

This guide walks you through preparing your Go environment, containerizing your Hyperledger Fabric external chaincode server, and publishing it to Docker Hub.

---

## 1. Prerequisites

Before starting, ensure you have the following installed and configured on your machine:

* **Go Lang**: Version 1.22 or higher installed locally.
* **Docker Desktop**: Installed, running, and configured on your machine (including WSL2 backing if you are on Windows).
* **Git**: (Optional) For version control tracking.

---

## 2. Set Up Your Docker Hub Account & Repository

### Step 2.1: Create a Docker Hub Account

If you do not have an account, head over to [Docker Hub](https://hub.docker.com/) and sign up for a free account. Change [USERNAME] to your docker hub username.

### Step 2.2: Create a Docker Repository

1. Log in to your Docker Hub dashboard.
2. Click on the **Create Repository** button.
3. Set the repository name. Change below [REPOSITORY_NAME] to your repository name you set here. 
4. Choose **Public** for MBI usage.
5. Click **Create**.

Your full image reference tag will look like this: `[USERNAME]/[REPOSITORY_NAME]:<tag>`

---

## 3. Authenticate Locally (Docker Login)

Open your terminal (PowerShell, Command Prompt, or Bash) and run the login command to authenticate your local Docker client with your remote account:

```bash
docker login -u [USERNAME]

```

> 💡 **Security Tip:** When prompted for a password, it is highly recommended to use a **Personal Access Token (PAT)** generated from your Docker Hub Account Settings instead of your raw password.

---

## 4. Prepare the Go Project Dependencies

To prevent the Docker build from failing due to a missing `go.sum` file, generate the dependency checksums locally in your project root (`asset-registry` folder):

```bash
go mod tidy

```

This command automatically downloads the required Hyperledger Fabric `v2` dependencies and generates a verified `go.sum` file next to your `go.mod`.

---

## 5. Build the Docker Image

With your workspace prepared, execute the Docker build command. We will tag this version as `0.1`:

```bash
docker build -t [USERNAME]/[REPOSITORY_NAME]:0.1 .

```

* `-t`: Applies a name and an optional tag in the `name:tag` format.
* `.`: Tells Docker to look for the `Dockerfile` inside your current working directory.

---

## 6. Push the Image to Docker Hub

Now that the image is built locally, upload it directly to your general repository branch on Docker Hub:

```bash
docker push [USERNAME]/[REPOSITORY_NAME]:0.1

```