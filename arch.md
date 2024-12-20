# Arch Diagram

┌─────────────────┐          ┌─────────────────┐
│   HTMX Frontend │◄────────►│   Go Backend    │
└─────────────────┘          └────────┬────────┘
                                      │
                             ┌────────┼────────┐
                             │        │        │
                       ┌─────▼─┐  ┌──▼───┐   ┌─▼────┐
                       │ Redis │  │WebSkt│   │Memory│
                       │Session│  │Server│   │ Quiz │
                       └───────┘  └──────┘   └──────┘

Right now, the backend is a simple Go server that uses the stdlib `net/http` package to serve the HTMX frontend. The frontend is a single-page application that uses HTMX to communicate with the backend.

There is 3 main components on server side:
- Quiz Session: 
  This is the main component that manages the quiz session. The score calculation and leaderboard updates are handled here.
- WebSocket Server: 
  This is the component that handles the WebSocket connection between the client and the server for leaderboard updates.
- Memory: 
  This is the component that manages the in-memory quizes.

## Infrastructure

The infrastructure is built using the following technologies:
- Go: 
  The backend is written in Go, which is a statically typed, compiled language that is easy to learn and use.
- Redis: 
  The backend uses Redis as a session store and a cache for frequently accessed quizzes.
- WebSocket: 
  The backend uses WebSocket to handle real-time updates from the client.
- HTMX: 
  The frontend is built using HTMX, which is a JavaScript library that makes it easy to build interactive web applications.
- Docker: 
  The backend is containerized using Docker, which allows for easy deployment and scaling.  
- Docker Compose: 
  The backend is configured using Docker Compose, which simplifies the setup process.


## Future Development

For future development, I would like to add the following components:
1. Quiz Management System:
   This is the component that manages the quiz creation, deletion, and modification.
   With this component, I can:
    - Add quiz categories and difficulty levels
    - Support multiple question types (MCQ, True/False, Fill-in-blanks)
    - Quiz versioning system
    - Quiz templates and cloning functionality
    - Quiz scheduling system
2. Storage Improvements: 
    - Migrate quiz storage from in-memory to persistent database
    - Implement caching layer for frequently accessed quizzes
3. Real-time Features:
    - Chat system between players
    - Spectator mode
    - Live quiz creation/modification
    - Real-time analytics dashboard
    - Interactive elements (polls, reactions)
4. Analytics and Reporting:
    - Player performance analytics
    - Quiz difficulty analysis
    - Response time tracking
    - Success rate metrics
    - Export functionality for results
5. Game Mechanics:
    - Power-ups and special abilities
    - Time-based scoring
    - Team-based competitions
    - Tournament mode
    - Practice mode