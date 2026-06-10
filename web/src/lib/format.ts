export function execFormat(command: string, value?: string): void {
  try {
    document.execCommand(command, false, value);
  } catch (e) {
    console.warn(`execCommand('${command}') failed:`, e);
  }
}

export function queryFormatState(command: string): boolean {
  try {
    return document.queryCommandState(command);
  } catch {
    return false;
  }
}

export function queryFormatValue(command: string): string | null {
  try {
    return document.queryCommandValue(command);
  } catch {
    return null;
  }
}
