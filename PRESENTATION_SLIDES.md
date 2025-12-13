# Remote Shell RPC System - Presentation Slides

**Duration**: 5 minutes  
**Group ID**: [GID]  
**Date**: December 16, 2025

---

## Slide 1: Title Slide
```
╔═══════════════════════════════════════════════════╗
║                                                   ║
║     Remote Shell RPC System                      ║
║     Distributed Systems Midterm Project          ║
║                                                   ║
║     Group: [GID]                                 ║
║     Date: December 16, 2025                       ║
║                                                   ║
╚═══════════════════════════════════════════════════╝
```

**Speaker Notes:**
- Introduce project title
- Mention it's a distributed systems project
- State group ID

---

## Slide 2: Project Overview
```
┌─────────────────────────────────────────────────┐
│  What is Remote Shell RPC System?              │
│                                                 │
│  • RPC-based remote shell execution            │
│  • Multiple clients connect to one server      │
│  • Similar to kubectl exec in Kubernetes       │
│  • Cross-platform (Windows, Linux, macOS)      │
│                                                 │
│  Key Features:                                  │
│  ✓ Concurrent client handling                   │
│  ✓ Session management                           │
│  ✓ Interactive & non-interactive modes          │
│  ✓ Client tracking                              │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Explain what the system does
- Highlight key features
- Mention real-world application (kubectl exec)

---

## Slide 3: Architecture Design
```
┌─────────────────────────────────────────────────┐
│  System Architecture                            │
│                                                 │
│         ┌──────────────┐                       │
│         │ RPC Server   │                       │
│         │  :8080       │                       │
│         └──────┬───────┘                       │
│                │                               │
│        ┌───────┼───────┐                      │
│        │       │       │                       │
│   ┌────▼──┐ ┌─▼───┐ ┌─▼───┐                  │
│   │Client1│ │Clnt2│ │Clnt3│                  │
│   └───────┘ └─────┘ └─────┘                  │
│                                                 │
│  • Star topology                                │
│  • TCP/IP protocol                              │
│  • Goroutines for concurrency                   │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Explain star topology
- Show how multiple clients connect
- Mention concurrent processing with goroutines

---

## Slide 4: Protocol Design
```
┌─────────────────────────────────────────────────┐
│  RPC Communication Protocol                     │
│                                                 │
│  1. Connection: TCP on port 8080                │
│  2. Registration: Client → Server              │
│     Register(clientID)                          │
│                                                 │
│  3. Command Execution:                          │
│     Client → Execute(command) → Server          │
│     Server → Response(output, exitCode)          │
│                                                 │
│  4. Session Management:                        │
│     • Working directory per client              │
│     • Environment variables per client          │
│     • Thread-safe with mutex                    │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Explain RPC protocol
- Show communication flow
- Mention session management

---

## Slide 5: Key RPC Methods
```
┌─────────────────────────────────────────────────┐
│  RPC Methods                                    │
│                                                 │
│  1. Register(clientID)                          │
│     → Creates session for client                │
│                                                 │
│  2. Execute(command)                            │
│     → Runs shell command                        │
│     → Returns output + exit code                │
│                                                 │
│  3. SetEnv(key, value)                          │
│     → Sets environment variable                 │
│                                                 │
│  4. ChangeDir(directory)                        │
│     → Changes working directory                 │
│                                                 │
│  5. ListClients()                               │
│     → Returns all active clients                │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- List all RPC methods
- Explain what each does
- Show their purpose

---

## Slide 6: Deployment
```
┌─────────────────────────────────────────────────┐
│  Deployment Steps                               │
│                                                 │
│  1. Install Go (1.21+)                          │
│  2. Clone/download project                      │
│  3. Build:                                      │
│     .\build.ps1  (Windows)                     │
│     make build   (Linux/Mac)                    │
│                                                 │
│  4. Run Server:                                 │
│     .\bin\server.exe                            │
│                                                 │
│  5. Run Clients:                                │
│     .\bin\client.exe -id client1               │
│     .\bin\client.exe -id client2               │
│                                                 │
│  6. Check Status:                               │
│     .\bin\admin.exe                             │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Quick deployment overview
- Show it's easy to set up
- Mention cross-platform support

---

## Slide 7: Demo - Multiple Clients
```
┌─────────────────────────────────────────────────┐
│  Demo: Multiple Clients                         │
│                                                 │
│  Terminal 1 (Server):                          │
│  > Remote Shell RPC Server started             │
│  > [Client client1] Registered                  │
│  > [Client client2] Registered                  │
│                                                 │
│  Terminal 2 (Client 1):                        │
│  > [client1@remote]$ dir                        │
│  > [output...]                                  │
│                                                 │
│  Terminal 3 (Client 2):                        │
│  > [client2@remote]$ echo Hello                 │
│  > Hello                                        │
│                                                 │
│  Terminal 4 (Admin):                           │
│  > Active clients (2):                          │
│  >   1. client1                                 │
│  >   2. client2                                 │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Show live demo if possible
- Or show screenshots
- Demonstrate concurrent execution

---

## Slide 8: Results & Features
```
┌─────────────────────────────────────────────────┐
│  Results & Key Features                         │
│                                                 │
│  ✅ Multiple concurrent clients                 │
│  ✅ Session isolation                           │
│  ✅ Working directory per session               │
│  ✅ Environment variables per session            │
│  ✅ Interactive & non-interactive modes          │
│  ✅ Client tracking                             │
│  ✅ Cross-platform support                      │
│  ✅ Efficient with goroutines                   │
│                                                 │
│  Performance:                                   │
│  • Handles 10+ concurrent clients              │
│  • Low latency command execution                │
│  • Thread-safe operations                       │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Highlight achievements
- Show what works
- Mention performance

---

## Slide 9: Challenges & Solutions
```
┌─────────────────────────────────────────────────┐
│  Challenges & Solutions                         │
│                                                 │
│  Challenge 1: Concurrent Access                │
│  Solution: Mutex for thread-safe operations    │
│                                                 │
│  Challenge 2: Session Management               │
│  Solution: Map with client ID as key           │
│                                                 │
│  Challenge 3: Cross-platform Commands           │
│  Solution: Runtime detection (Windows vs Unix) │
│                                                 │
│  Challenge 4: Client Tracking                  │
│  Solution: Register on connection               │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Discuss challenges faced
- Explain solutions
- Show problem-solving skills

---

## Slide 10: Future Improvements
```
┌─────────────────────────────────────────────────┐
│  Future Improvements                           │
│                                                 │
│  • Real-time streaming output                   │
│  • Interactive TTY support                      │
│  • Authentication & authorization               │
│  • TLS/SSL encryption                           │
│  • Command history                              │
│  • File transfer (scp-like)                    │
│  • gRPC instead of net/rpc                      │
│  • Metrics & monitoring dashboard               │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- Mention future enhancements
- Show understanding of limitations
- Demonstrate forward thinking

---

## Slide 11: Team Contributions
```
┌─────────────────────────────────────────────────┐
│  Team Contributions                             │
│                                                 │
│  Member 1: [Name]                               │
│  • Server implementation (30%)                  │
│  • Session management (15%)                     │
│  • Total: 45%                                   │
│                                                 │
│  Member 2: [Name]                               │
│  • Client implementation (25%)                  │
│  • Admin tool (10%)                             │
│  • Testing (10%)                                │
│  • Total: 45%                                   │
│                                                 │
│  Member 3: [Name]                               │
│  • Documentation (10%)                           │
│  • Total: 10%                                   │
└─────────────────────────────────────────────────┘
```

**Speaker Notes:**
- List each member's contribution
- Show fair distribution
- Be ready for questions

---

## Slide 12: Q&A
```
╔═══════════════════════════════════════════════════╗
║                                                   ║
║              Thank You!                           ║
║                                                   ║
║              Questions?                          ║
║                                                   ║
║     GitHub: [repository URL]                     ║
║     Email: [group email]                         ║
║                                                   ║
╚═══════════════════════════════════════════════════╝
```

**Speaker Notes:**
- Thank audience
- Invite questions
- Provide contact info

---

## Presentation Tips

### Timing (5 minutes)
- **Slide 1-2**: 30 seconds (Introduction)
- **Slide 3-5**: 2 minutes (Architecture & Protocol)
- **Slide 6-7**: 1 minute (Deployment & Demo)
- **Slide 8-10**: 1 minute (Results & Future)
- **Slide 11**: 30 seconds (Contributions)
- **Slide 12**: 30 seconds (Q&A)

### Preparation Tips
1. **Practice**: Rehearse 2-3 times
2. **Demo**: Prepare live demo or screenshots
3. **Q&A**: Prepare answers for common questions:
   - "How does RPC work?"
   - "How do you handle concurrent access?"
   - "What are the limitations?"
   - "How is this different from SSH?"

### Common Questions & Answers

**Q: Why RPC instead of REST API?**
A: RPC is more efficient for remote procedure calls, better suited for this use case. Lower overhead than HTTP/REST.

**Q: How do you ensure thread-safety?**
A: We use mutex (sync.Mutex) to protect shared data structures (sessions map) from concurrent access.

**Q: What happens if a client disconnects?**
A: The session remains in memory until server restarts. Could add cleanup in future.

**Q: Can this scale to 100+ clients?**
A: Yes, goroutines are lightweight. The main limitation is network bandwidth, not server capacity.

**Q: How is this different from SSH?**
A: SSH requires authentication, this is simpler for internal use. SSH is more secure but heavier.

---

## Slide Design Tips

1. **Keep it simple**: Max 5-7 bullet points per slide
2. **Use visuals**: Diagrams, screenshots, code snippets
3. **Consistent style**: Same font, colors, layout
4. **Readable**: Large fonts, good contrast
5. **Practice**: Time yourself, know your content

---

**Good luck with your presentation!**



