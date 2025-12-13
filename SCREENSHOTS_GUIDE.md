# Screenshots Guide - How to Capture Results

This guide helps you take all required screenshots for the midterm report.

---

## Required Screenshots (8 minimum)

### 1. Server Startup
**What to capture:**
- Server terminal showing "Remote Shell RPC Server started on :8080"
- "Waiting for clients..." message

**How to take:**
1. Open PowerShell/Terminal
2. Navigate to project directory
3. Run: `.\bin\server.exe` (or `./bin/server` on Linux/Mac)
4. Take screenshot of the terminal

**Example:**
```
Remote Shell RPC Server started on :8080
Waiting for clients...
```

---

### 2. Client Connection
**What to capture:**
- Client terminal showing successful connection
- "Connected to server localhost:8080 as [client-id]" message
- Prompt showing `[client-id@remote]$`

**How to take:**
1. Open new terminal
2. Run: `.\bin\client.exe -id client1`
3. Wait for connection message
4. Take screenshot

**Example:**
```
Connected to server localhost:8080 as client1
Type 'exit' to quit, 'help' for commands
[client1@remote]$ 
```

---

### 3. Command Execution
**What to capture:**
- Client terminal with command input
- Command output showing results

**How to take:**
1. In client terminal, type a command (e.g., `dir` on Windows or `ls` on Linux)
2. Press Enter
3. Wait for output
4. Take screenshot showing both command and output

**Example (Windows):**
```
[client1@remote]$ dir
 Volume in drive C has no label.
 Directory of C:\Users\Admin\Downloads\abcd

[client1@remote]$ 
```

**Example (Linux/Mac):**
```
[client1@remote]$ ls -la
total 24
drwxr-xr-x  3 user user 4096 Jan 15 10:00 .
[client1@remote]$ 
```

---

### 4. Multiple Clients
**What to capture:**
- Multiple terminal windows showing different clients
- At least 2 clients connected simultaneously

**How to take:**
1. Open Terminal 1: Run server
2. Open Terminal 2: Run `.\bin\client.exe -id client1`
3. Open Terminal 3: Run `.\bin\client.exe -id client2`
4. Arrange terminals side by side
5. Take screenshot showing all 3 terminals

**Layout suggestion:**
```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   Server    │  │  Client 1   │  │  Client 2   │
│             │  │             │  │             │
│  Started    │  │ [client1@   │  │ [client2@   │
│  Waiting... │  │  remote]$   │  │  remote]$   │
└─────────────┘  └─────────────┘  └─────────────┘
```

---

### 5. Admin Tool - List Clients
**What to capture:**
- Admin terminal showing list of active clients
- Should show 2+ clients

**How to take:**
1. With 2+ clients connected (from screenshot 4)
2. Open new terminal
3. Run: `.\bin\admin.exe`
4. Take screenshot showing the list

**Example:**
```
Active clients (2):
  1. client1
  2. client2
```

---

### 6. Change Directory
**What to capture:**
- Client using `cd` command
- Directory change confirmation
- Listing files in new directory

**How to take:**
1. In client terminal, type: `cd C:\Windows` (or `/tmp` on Linux)
2. Press Enter
3. Type: `dir` (or `ls`)
4. Take screenshot showing all steps

**Example:**
```
[client1@remote]$ cd C:\Windows
Directory changed
[client1@remote]$ dir
...
[files listed]
[client1@remote]$ 
```

---

### 7. Environment Variables
**What to capture:**
- Setting environment variable with `setenv`
- Using the variable (e.g., `echo %VAR%` on Windows or `echo $VAR` on Linux)

**How to take:**
1. In client terminal, type: `setenv MY_VAR hello`
2. Press Enter
3. Type: `echo %MY_VAR%` (Windows) or `echo $MY_VAR` (Linux)
4. Take screenshot showing both commands and output

**Example (Windows):**
```
[client1@remote]$ setenv MY_VAR hello
Set MY_VAR=hello
[client1@remote]$ echo %MY_VAR%
hello
[client1@remote]$ 
```

**Example (Linux/Mac):**
```
[client1@remote]$ setenv MY_VAR hello
Set MY_VAR=hello
[client1@remote]$ echo $MY_VAR
hello
[client1@remote]$ 
```

---

### 8. Server Logs
**What to capture:**
- Server terminal showing logs from multiple clients
- Registration messages
- Command execution logs

**How to take:**
1. With 2+ clients connected and executing commands
2. Go to server terminal
3. Scroll to show recent logs
4. Take screenshot

**Example:**
```
Remote Shell RPC Server started on :8080
Waiting for clients...
New client connected: 127.0.0.1:xxxxx
[Client client1] Registered (new session)
[Client client1] Executed: dir (Exit: 0)
New client connected: 127.0.0.1:xxxxx
[Client client2] Registered (new session)
[Client client2] Executed: echo Hello (Exit: 0)
```

---

## Screenshot Tips

### Quality
- **Resolution**: At least 1920x1080 or higher
- **Format**: PNG (best quality) or JPG
- **Size**: Keep file size reasonable (< 2MB each)

### Clarity
- **Font size**: Make terminal font large enough to read
- **Zoom**: Zoom in if text is too small
- **Focus**: Ensure text is sharp and readable

### Organization
- **Naming**: Use descriptive names:
  - `01-server-startup.png`
  - `02-client-connection.png`
  - `03-command-execution.png`
  - etc.

- **Folder**: Create `screenshots/` folder
- **Order**: Number them in sequence

### Tools

**Windows:**
- `Win + Shift + S`: Snipping Tool (select area)
- `Alt + Print Screen`: Screenshot active window
- `Win + Print Screen`: Full screen

**Linux:**
- `Print Screen`: Full screen
- `Shift + Print Screen`: Select area
- `gnome-screenshot`: GUI tool

**macOS:**
- `Cmd + Shift + 4`: Select area
- `Cmd + Shift + 3`: Full screen
- `Cmd + Shift + 4 + Space`: Window screenshot

---

## Screenshot Workflow

### Step 1: Prepare
1. Close unnecessary applications
2. Clean up terminal windows
3. Set terminal font size to readable (14-16pt)
4. Arrange windows nicely

### Step 2: Execute
1. Start server
2. Connect clients
3. Execute commands
4. Run admin tool

### Step 3: Capture
1. Take screenshots in order
2. Save with descriptive names
3. Verify all screenshots are clear

### Step 4: Insert
1. Insert screenshots into report
2. Add captions: "Figure 1: Server startup"
3. Reference in text: "As shown in Figure 1..."

---

## Example Screenshot Layout for Report

```markdown
## 6. Results Captured as Images

### 6.1 Server Startup
![Server Startup](screenshots/01-server-startup.png)
*Figure 1: RPC Server starting and waiting for clients*

### 6.2 Client Connection
![Client Connection](screenshots/02-client-connection.png)
*Figure 2: Client successfully connecting to server*

### 6.3 Command Execution
![Command Execution](screenshots/03-command-execution.png)
*Figure 3: Executing 'dir' command remotely*

... (continue for all 8 screenshots)
```

---

## Checklist

Before submitting, verify:
- [ ] All 8 screenshots taken
- [ ] Screenshots are clear and readable
- [ ] Text in screenshots is large enough
- [ ] Screenshots show the required functionality
- [ ] Screenshots are inserted into report
- [ ] Each screenshot has a caption
- [ ] Screenshots are referenced in text

---

## Troubleshooting

**Problem: Screenshot too small**
- Solution: Zoom in terminal or increase font size

**Problem: Text not readable**
- Solution: Increase terminal font size, use high DPI

**Problem: Too much clutter**
- Solution: Close unnecessary windows, focus on relevant terminals

**Problem: Can't capture multiple terminals**
- Solution: Use window manager to tile windows, or take separate screenshots

---

**Good luck with your screenshots! 📸**



