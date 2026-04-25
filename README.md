# Workspace MCP

An MCP server that allows AI agents to access Google documents and spreadsheets.

## Tools

### Docs

| Tool | Description |
|------|-------------|
| `GetDocumentSize` | Returns the character count of a Google Doc |
| `ReadDocumentAsMarkdown` | Fetches a Google Doc as markdown, with character-based pagination |

### Sheets

| Tool | Description |
|------|-------------|
| `ListSheets` | Returns the names of all sheets in a spreadsheet |
| `GetSheetSize` | Returns the row/column extent of actual data in a sheet |
| `ReadSheet` | Reads a range of cells and returns them as a 2D array |

## Setup

### Setting up Google OAuth

This app uses OAuth 2.0 to authenticate with Google and access your Docs and Sheets on your behalf. OAuth is a standard protocol that lets you grant the app access to your files without sharing your Google password. This setup involves creating a Google Cloud project and configuring credentials — it takes about 10–15 minutes and is the most involved part of the process.

#### 1. Create a Google Cloud project

1. Go to [console.cloud.google.com](https://console.cloud.google.com) and sign in
2. Click the project dropdown in the top navigation bar → **New Project**
3. Enter a project name (e.g. `workspace-mcp`) and click **Create**
4. Wait a moment, then make sure the new project is selected in the dropdown

#### 2. Enable the required APIs

1. In the left sidebar, go to **APIs & Services → Library**
2. Search for **Google Drive API**, click on it, then click **Enable**
3. Go back to the library, search for **Google Sheets API**, click on it, then click **Enable**

#### 3. Configure the OAuth consent screen

The consent screen is what Google shows users when they authorize an app. Even for personal use, it must be configured.

1. In the left sidebar, go to **APIs & Services → OAuth consent screen**
2. If prompted with "Google Auth Platform not configured yet", click **Get Started**
3. Under **App Information**, fill in:
   - **App name**: anything, e.g. `Workspace MCP`
   - **User support email**: your email address
4. Click **Next**
5. Under **Audience**, select **External**, then click **Next**
6. Under **Contact Information**, enter your email address, then click **Next**
7. Check the box to agree to the Google API Services User Data Policy, then click **Continue** → **Create**

#### 4. Add OAuth scopes

Scopes define exactly what data the app is allowed to access. You need to add the two scopes this app uses.

1. In the left sidebar, go to **APIs & Services → OAuth consent screen** (or **Google Auth Platform → Data Access**)
2. Click **Add or Remove Scopes**
3. In the search box, paste `https://www.googleapis.com/auth/drive.readonly` and press Enter, then check the box next to it
4. Clear the search box, paste `https://www.googleapis.com/auth/spreadsheets.readonly` and press Enter, then check the box next to it
5. Click **Update**, then **Save and Continue**

#### 5. Add yourself as a test user

Since the app is not verified by Google, only explicitly added test users can authorize it.

1. In the left sidebar, go to **APIs & Services → OAuth consent screen**
2. Navigate to the **Audience** section
3. Under **Test users**, click **Add Users**
4. Enter your Google account email and click **Save**

> Without this step, you will see a "Access blocked" error when trying to authorize. If the token expires and you need to re-authorize, this step is already done.

#### 6. Create OAuth credentials

1. In the left sidebar, go to **APIs & Services → Credentials**
2. Click **Create Credentials → OAuth client ID**
3. For **Application type**, select **Desktop app**
4. Give it a name (e.g. `workspace-mcp-client`) and click **Create**
5. In the dialog that appears, click **Download JSON**
6. Save the downloaded file somewhere on your machine, e.g. `~/workspace-mcp/credentials.json`

> This file contains your client ID and secret. Do not share it or commit it to version control.

---

### Create the config file

Create a JSON file anywhere on your machine, for example `~/workspace-mcp/config.json`:

```json
{
  "googleCredentialsFile": "/path/to/credentials.json",
  "googleTokenFile": "/path/to/token.json",
  "logFilePath": "/path/to/workspace-mcp.log",
  "oauthCallbackPort": 47291
}
```

| Field | Description |
|-------|-------------|
| `googleCredentialsFile` | Path to the credentials JSON downloaded in the OAuth setup |
| `googleTokenFile` | Where the server stores your OAuth token after first login (created automatically) |
| `logFilePath` | Where the server writes its logs |
| `oauthCallbackPort` | Local port used during the one-time OAuth login flow |

---

### Build the server

```bash
git clone https://github.com/shivanshkc/workspacemcp
cd workspacemcp
make build
```

The binary is written to `bin/workspace-mcp`.

---

### Register with Claude Code

Open the `Makefile` and update `default_config_path` to point to the config file you created above. Then run:

```bash
make claude
```

This builds the binary and registers it as an MCP server in Claude Code.

---

### Authorize with Google (first run only)

The next time Claude Code starts and connects to this server, it will trigger the OAuth flow automatically:

1. Your browser will open to a Google sign-in page
2. Sign in with the Google account you added as a test user
3. You may see a warning saying **"Google hasn't verified this app"** — click **Advanced** → **Go to [app name] (unsafe)** to proceed. This is expected for personal, unverified apps.
4. Grant the requested permissions
5. The browser will show **"Authorization complete. You may close this tab."**

The token is saved to `googleTokenFile` and reused on all subsequent runs. You will not need to authorize again unless you delete the token file.

> **Re-authorizing:** If you need to re-authorize (e.g. after changing scopes or if the token expires), delete the token file and restart Claude Code:
> ```bash
> rm /path/to/token.json
> ```

---

## License

[MIT](LICENSE)
