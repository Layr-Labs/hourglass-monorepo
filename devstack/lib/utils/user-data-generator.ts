import * as fs from 'fs';
import * as path from 'path';

export interface UserDataParams {
  forkUrl: string;
}

export function generateUserDataScript(params: UserDataParams): string {
  // Read the template script
  const templatePath = path.join(__dirname, '..', 'config', 'user-data.sh');
  let script = fs.readFileSync(templatePath, 'utf8');

  // Replace placeholders with actual values
  const replacements: Record<string, string> = {
    'FORK_URL_PLACEHOLDER': params.forkUrl || '',
  };

  for (const [placeholder, value] of Object.entries(replacements)) {
    script = script.replace(new RegExp(placeholder, 'g'), value);
  }

  return script;
}