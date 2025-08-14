#!/usr/bin/env node

const { execSync, spawn } = require("child_process");
const path = require("path");
const fs = require("fs");

const extractDir = path.join(__dirname, "dist");

// ensure output dir
fs.mkdirSync(extractDir, { recursive: true });

function replaceEnv() {
    // copy .env file to dist
    fs.copyFileSync(".env", path.join(extractDir, ".env"));
}

function createDatabaseAndRunMigrations() {
    // updating
}

function extractAndRun(baseName, launch) {
    const binaryFile = path.join(extractDir, `${baseName}`);
    if (!fs.existsSync(binaryFile)) {
        console.error(`❌ Binary file not found: ${binaryFile}`);
        process.exit(1);
    }

    launch(baseName, extractDir);
}

console.log(`📦 Extracting package...`);

replaceEnv();

extractAndRun("./worker", (bin, workingDir) => {
    console.log(`🚀 Spawn process and run worker...`);
    const proc = spawn(bin, [], { stdio: ["pipe", "pipe", "pipe"], cwd: workingDir });
    process.stdin.pipe(proc.stdin);
    proc.stdout.pipe(process.stdout);
    proc.stderr.pipe(process.stdout);

    proc.on("exit", (c) => process.exit(c || 0));
    proc.on("error", (e) => {
        console.error("❌ Worker error:", e.message);
        process.exit(1);
    });
    process.on("SIGINT", () => {
        console.error("\n🛑 Shutting down worker...");
        proc.kill("SIGINT");
    });
    process.on("SIGTERM", () => proc.kill("SIGTERM"));
});

extractAndRun("./server", (bin, workingDir) => {
    console.log(`🚀 Launching...`);
    // execSync(`"${bin}"`, { stdio: "inherit", cwd: workingDir });
    const proc = spawn(bin, [], { stdio: ["pipe", "pipe", "pipe"], cwd: workingDir });
    process.stdin.pipe(proc.stdin);
    proc.stdout.pipe(process.stdout);
    proc.stderr.pipe(process.stdout);

    proc.on("exit", (c) => process.exit(c || 0));
    proc.on("error", (e) => {
        console.error("❌ Server error:", e.message);
        process.exit(1);
    });
    process.on("SIGINT", () => {
        console.error("\n🛑 Shutting down server...");
        proc.kill("SIGINT");
    });
    process.on("SIGTERM", () => proc.kill("SIGTERM"));
});