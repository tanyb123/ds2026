# Remote Shell RPC System - Architecture & ERM

## Entity Relationship Model (ERM)

```mermaid
erDiagram
    SERVER ||--o{ SESSION : manages
    SESSION ||--o{ ENV_VAR : contains
    CLIENT ||--|| SESSION : uses
    SERVER ||--o{ CONNECTION : accepts
    CONNECTION ||--|| CLIENT : represents
    
    SERVER {
        string address
        int port
        map sessions
        mutex lock
    }
    
    SESSION {
        string id
        string workDir
        datetime connectedAt
        datetime lastActive
        bool isActive
    }
    
    ENV_VAR {
        string key
        string value
        string sessionId
    }
    
    CLIENT {
        string id
        string serverAddr
        bool connected
        datetime lastHeartbeat
    }
    
    CONNECTION {
        string remoteAddr
        datetime establishedAt
        datetime lastActivity
    }
    
    COMMAND_REQUEST {
        string command
        array args
        string clientId
    }
    
    COMMAND_RESPONSE {
        string output
        string error
        int exitCode
        string clientId
    }
    
    SESSION ||--o{ COMMAND_REQUEST : receives
    SESSION ||--o{ COMMAND_RESPONSE : sends
```

## System Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        C1[Client 1]
        C2[Client 2]
        C3[Client N]
        ADMIN[Admin Tool]
    end
    
    subgraph "Network Layer"
        TCP[TCP Connection]
        RPC[RPC Protocol]
    end
    
    subgraph "Server Layer"
        SERVER[RPC Server]
        SERVICE[RemoteShellService]
        SESSIONS[Session Manager]
        CLEANUP[Cleanup Goroutine]
    end
    
    subgraph "Session Storage"
        S1[Session 1<br/>ID, Env, WorkDir]
        S2[Session 2<br/>ID, Env, WorkDir]
        SN[Session N<br/>ID, Env, WorkDir]
    end
    
    subgraph "Execution Layer"
        EXEC[Command Executor]
        TIMEOUT[Timeout Handler]
        ENV[Environment Manager]
    end
    
    C1 -->|RPC Call| TCP
    C2 -->|RPC Call| TCP
    C3 -->|RPC Call| TCP
    ADMIN -->|Admin RPCs<br/>(list / sessions / kill / whitelist)| TCP
    
    TCP --> RPC
    RPC --> SERVER
    
    SERVER --> SERVICE
    SERVICE --> SESSIONS
    SERVICE --> EXEC
    
    SESSIONS --> S1
    SESSIONS --> S2
    SESSIONS --> SN
    
    SESSIONS --> CLEANUP
    CLEANUP -.->|Remove Inactive| S1
    CLEANUP -.->|Remove Inactive| S2
    CLEANUP -.->|Remove Inactive| SN
    
    EXEC --> TIMEOUT
    EXEC --> ENV
    ENV --> S1
    ENV --> S2
    ENV --> SN
    
    style SERVER fill:#4a90e2
    style SERVICE fill:#50c878
    style SESSIONS fill:#ff6b6b
    style EXEC fill:#feca57
```

## Sequence Diagram - Command Execution

```mermaid
sequenceDiagram
    participant C as Client
    participant S as RPC Server
    participant SM as Session Manager
    participant E as Executor
    
    C->>S: Execute(command, clientID)
    S->>SM: Get/Create Session(clientID)
    SM-->>S: Session
    S->>E: Execute Command(command, session)
    E->>E: Set Timeout (5 min)
    E->>E: Set Working Directory
    E->>E: Set Environment Variables
    E->>E: Run Shell Command
    E-->>S: CommandResponse
    S->>SM: Update LastActive(clientID)
    S-->>C: CommandResponse
    
    Note over C,S: Heartbeat every 1 minute
    C->>S: Heartbeat(clientID)
    S->>SM: Update LastActive(clientID)
    S-->>C: OK
```

## Sequence Diagram - Session Cleanup

```mermaid
sequenceDiagram
    participant CG as Cleanup Goroutine
    participant SM as Session Manager
    participant S1 as Session 1 (Active)
    participant S2 as Session 2 (Inactive)
    
    Note over CG: Every 5 minutes
    CG->>SM: Check All Sessions
    SM->>S1: Check LastActive
    S1-->>SM: LastActive < 30 min (Active)
    SM->>S2: Check LastActive
    S2-->>SM: LastActive > 30 min (Inactive)
    SM->>S2: Delete Session
    Note over S2: Session Removed
```

## Component Interaction

```mermaid
graph LR
    subgraph "Concurrency Control"
        MUTEX[RWMutex Lock]
    end
    
    subgraph "Session Management"
        MAP[Session Map]
        REGISTER[Register]
        CLEANUP[Cleanup]
        HEARTBEAT[Heartbeat]
    end
    
    subgraph "RPC Methods"
        EXEC[Execute]
        SETENV[SetEnv]
        CHDIR[ChangeDir]
        LIST[ListClients]
        LISTSESS[ListSessions]
        KILL[KillSession]
        WL[AddToWhitelist]
    end
    
    subgraph "Error Handling"
        RECONNECT[Reconnect]
        TIMEOUT[Timeout]
        RETRY[Retry]
    end
    
    MUTEX --> MAP
    REGISTER --> MAP
    CLEANUP --> MAP
    HEARTBEAT --> MAP
    
    EXEC --> MUTEX
    SETENV --> MUTEX
    CHDIR --> MUTEX
    LIST --> MUTEX
    LISTSESS --> MUTEX
    KILL --> MUTEX
    WL --> MUTEX
    
    EXEC --> TIMEOUT
    RECONNECT --> RETRY
    
    style MUTEX fill:#ff6b6b
    style MAP fill:#4a90e2
    style TIMEOUT fill:#feca57
```

## Data Flow

```mermaid
flowchart TD
    START[Client Request] --> VALIDATE{Validate Request}
    VALIDATE -->|Invalid| ERROR[Return Error]
    VALIDATE -->|Valid| LOCK[Acquire Lock]
    
    LOCK --> SESSION{Session Exists?}
    SESSION -->|No| CREATE[Create Session]
    SESSION -->|Yes| UPDATE[Update LastActive]
    
    CREATE --> PREPARE[Prepare Command]
    UPDATE --> PREPARE
    
    PREPARE --> SETENV[Set Environment]
    SETENV --> SETDIR[Set Working Directory]
    SETDIR --> EXEC[Execute Command]
    
    EXEC --> TIMEOUT{Timeout?}
    TIMEOUT -->|Yes| TIMEOUT_ERR[Return Timeout Error]
    TIMEOUT -->|No| RESULT[Get Result]
    
    RESULT --> UNLOCK[Release Lock]
    TIMEOUT_ERR --> UNLOCK
    UNLOCK --> RESPONSE[Send Response]
    ERROR --> RESPONSE
    
    style START fill:#50c878
    style ERROR fill:#ff6b6b
    style RESPONSE fill:#4a90e2
```

## Key Features

### 1. Concurrency
- **Multiple Clients**: Server handles multiple clients simultaneously using goroutines
- **Thread Safety**: Uses `sync.RWMutex` to protect shared session data
- **Non-blocking**: Each client connection runs in separate goroutine

### 2. Fault Tolerance
- **Session Cleanup**: Automatic removal of inactive sessions (30 min timeout)
- **Reconnection**: Client can reconnect if connection is lost
- **Heartbeat**: Keepalive mechanism to detect dead connections
- **Timeout Handling**: Command execution timeout (5 minutes)

### 3. Session Management
- **Isolated Sessions**: Each client has independent session
- **Environment Variables**: Per-session environment variables
- **Working Directory**: Per-session working directory
- **Activity Tracking**: Last active time tracking

### 4. Error Handling
- **Connection Errors**: Automatic reconnection on client side
- **Command Timeout**: Prevents hanging commands
- **Graceful Degradation**: Server continues operating even if one client fails

