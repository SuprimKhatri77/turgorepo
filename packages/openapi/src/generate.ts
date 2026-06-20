import { writeFileSync, mkdirSync } from "fs";
import { dirname } from "path";
import { fileURLToPath } from "url";
import { generateOpenAPIDocument } from "./schema.ts";

const __dirname = dirname(fileURLToPath(import.meta.url));

const outPath = `${__dirname}/../../../apps/api/openapi.json`;

const doc = generateOpenAPIDocument();
const json = JSON.stringify(doc, null, 2);

mkdirSync(dirname(outPath), { recursive: true });
writeFileSync(outPath, json, "utf-8");

console.log("Generated openapi.json at", outPath);
