import { appendFile, chmod, mkdir } from "node:fs/promises";
import os from "node:os";
import path from "node:path";

export class TraceWriter {
  readonly filePath?: string;
  private pending: Promise<void> = Promise.resolve();

  constructor(filePath?: string) {
    this.filePath = filePath ? expandHome(filePath) : undefined;
  }

  write(record: Record<string, unknown>): Promise<void> {
    if (!this.filePath) return Promise.resolve();
    const line = `${JSON.stringify({ recorded_at: new Date().toISOString(), ...record })}\n`;
    this.pending = this.pending.catch(() => undefined).then(async () => {
      await mkdir(path.dirname(this.filePath!), { recursive: true, mode: 0o700 });
      await appendFile(this.filePath!, line, { encoding: "utf8", mode: 0o600 });
      await chmod(this.filePath!, 0o600);
    });
    return this.pending;
  }
}

function expandHome(value: string): string {
  if (value === "~") return os.homedir();
  if (value.startsWith(`~${path.sep}`)) return path.join(os.homedir(), value.slice(2));
  return value;
}
