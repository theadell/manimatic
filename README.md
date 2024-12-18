# Manimatic: Math Animations with Natural Language

Manimatic is an app for creating animations with [Manim](https://docs.manim.community/) (a popular engine for explanatory math videos). Instead of writing Python code manually, you can describe your animation in plain language, and the app generates the code for you, compiles it, and returns both the animation video and the code. An integrated editor lets you tweak the generated script and recompile with your changes.


---

## Features
- **Natural Language Animation Creation:** Generate Manim animations with ease.
- **Integrated Code Editor:** Edit and recompile generated Python scripts directly in the app.
- **Video Delivery:** Automatically compiles, uploads, and serves animations.

---

## Built With
- **Backend:** Golang
- **Frontend:** React
- **Infrastructure:** AWS Cloud Development Kit (CDK) in TypeScript for Infrastructure-as-Code (IaC).

### Architecture Overview
The app comprises:
1. **Prompt Handling API:** Receives prompts and generates Python scripts using GPT-4.
2. **Task Queue:** Scripts are pushed to an SQS queue for processing.
3. **Worker Instance:** Compiles Python code into video, uploads to S3, and returns signed URLs.
4. **Client Updates:** Uses Server-Sent Events (SSE) to push results (code + video) to the frontend.

![Architecture Diagram](./architecture-diagram.png)

---

## Getting Started Locally

### Prerequisites
- Docker and Docker Compose
- LocalStack for mocking AWS services
- OpenAI API Key

### Steps
1. Clone this repository:
   ```bash
   git clone https://github.com/theadell/manimatic
   ```
2. your OpenAI API key to a .env.local file:
    ```bash
    OPENAI_API_KEY=your-api-key
    ```
3. Start the backend and infrastructure with Docker Compose:
    ```bash
    docker-compose up
    ```
4. Run the frontend
    ```bash
    cd frontend
    npm install
    npm run dev
    ```

Access the app at http://localhost:5173
