import React, { useState, useEffect, useRef } from 'react';
import './index.css';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:3001';
const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:3001/ws';

function App() {
  const [username, setUsername] = useState('');
  const [enteredUsername, setEnteredUsername] = useState('');
  const [game, setGame] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const wsRef = useRef(null);
  const gameIdRef = useRef(null);

  useEffect(() => {
    fetchLeaderboard();
    const interval = setInterval(fetchLeaderboard, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  const fetchLeaderboard = async () => {
    try {
      const response = await fetch(`${API_URL}/api/leaderboard`);
      if (response.ok) {
        const data = await response.json();
        setLeaderboard(Array.isArray(data) ? data : []);
      } else {
        setLeaderboard([]);
      }
    } catch (error) {
      console.error('Error fetching leaderboard:', error);
      setLeaderboard([]);
    }
  };

  const connectWebSocket = () => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      return;
    }

    const ws = new WebSocket(WS_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setError('');
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      handleWebSocketMessage(data);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setError('Connection error. Please refresh the page.');
    };

    ws.onclose = () => {
      console.log('WebSocket closed');
      // Attempt to reconnect if game is active
      if (game && game.status === 'active' && gameIdRef.current) {
        setTimeout(() => {
          reconnectToGame();
        }, 1000);
      }
    };
  };

  const handleWebSocketMessage = (data) => {
    switch (data.type) {
      case 'waiting':
        setMessage(data.message);
        setGame(null);
        break;
      case 'gameState':
        setGame(data.game);
        gameIdRef.current = data.game.id;
        setMessage('');
        setError('');
        break;
      case 'playerDisconnected':
        setMessage(data.message);
        break;
      case 'playerReconnected':
        setMessage(`${data.username} reconnected!`);
        break;
      case 'error':
        setError(data.message);
        break;
      default:
        console.log('Unknown message type:', data.type);
    }
  };

  const handleJoin = (e) => {
    e.preventDefault();
    if (!enteredUsername.trim()) {
      setError('Please enter a username');
      return;
    }

    setUsername(enteredUsername.trim());
    setError('');
    setMessage('');
    connectWebSocket();

    // Send join message
    setTimeout(() => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({
          type: 'join',
          username: enteredUsername.trim(),
        }));
      }
    }, 100);
  };

  const reconnectToGame = () => {
    if (!gameIdRef.current || !username) return;

    connectWebSocket();
    setTimeout(() => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({
          type: 'rejoin',
          username: username,
          gameId: gameIdRef.current,
        }));
      }
    }, 100);
  };

  const handleColumnClick = (column) => {
    if (!game || game.status !== 'active') return;
    
    // Check if it's player's turn (currentPlayer is now username)
    const isPlayerTurn = game.currentPlayer === username;
    
    if (!isPlayerTurn) return;

    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'makeMove',
        gameId: game.id,
        column: column,
      }));
    }
  };

  const getCellColor = (cell, rowIndex, colIndex) => {
    if (!cell || !game) return '';
    
    // Cell contains username, determine which player
    if (cell === game.player1.username) {
      return 'red';
    } else if (cell === game.player2.username) {
      return 'yellow';
    }
    return '';
  };

  const getStatusClass = () => {
    if (!game) return 'waiting';
    if (game.status === 'active') return 'active';
    return 'finished';
  };

  const getStatusMessage = () => {
    if (!game) {
      return message || 'Enter your username to start';
    }
    if (game.status === 'finished') {
      if (game.winner === 'draw') {
        return 'Game ended in a draw!';
      }
      const winnerName = game.winner === game.player1.username 
        ? game.player1.username 
        : game.player2.username;
      if (winnerName === username) {
        return 'You won! 🎉';
      }
      return `${winnerName} won!`;
    }
    // currentPlayer is now username
    if (game.currentPlayer === username) {
      return `Your turn! Drop a disc in a column.`;
    }
    
    if (game.player2.isBot && game.currentPlayer === game.player2.username) {
      return `Waiting for Bot...`;
    }
    
    return `Waiting for ${game.currentPlayer}...`;
  };

  return (
    <div className="app">
      <div className="header">
        <h1>🎮 4 in a Row - Connect Four</h1>
      </div>

      <div className="game-container">
        <div className="game-board-container">
          {!username ? (
            <div className="username-form">
              <h3>Enter Your Username</h3>
              <form onSubmit={handleJoin}>
                <input
                  type="text"
                  value={enteredUsername}
                  onChange={(e) => setEnteredUsername(e.target.value)}
                  placeholder="Username"
                  maxLength={20}
                />
                <button type="submit">Join Game</button>
              </form>
            </div>
          ) : (
            <>
              <div className="game-info">
                <h3>Game Info</h3>
                <p><strong>You:</strong> {username}</p>
                {game && (
                  <>
                    <p><strong>Opponent:</strong> {game.player2.username}</p>
                    <div className={`status ${getStatusClass()}`}>
                      {getStatusMessage()}
                    </div>
                  </>
                )}
                {error && <div className="error">{error}</div>}
                {message && !game && <div className="message">{message}</div>}
              </div>

              {game && (
                <div className="board">
                  <div className="column-buttons">
                    {[0, 1, 2, 3, 4, 5, 6].map((col) => (
                      <button
                        key={col}
                        className="column-button"
                        onClick={() => handleColumnClick(col)}
                        disabled={
                          game.status !== 'active' ||
                          game.currentPlayer !== username
                        }
                      >
                        ↓
                      </button>
                    ))}
                  </div>
                  {game.board && game.board.map((row, rowIndex) => (
                    <div key={rowIndex} className="board-row">
                      {row && row.map((cell, colIndex) => (
                        <div
                          key={`${rowIndex}-${colIndex}`}
                          className={`cell ${getCellColor(cell, rowIndex, colIndex)}`}
                        />
                      ))}
                    </div>
                  ))}
                </div>
              )}
            </>
          )}
        </div>

        <div className="sidebar">
          <div className="leaderboard">
            <h3>🏅 Leaderboard</h3>
            <table className="leaderboard-table">
              <thead>
                <tr>
                  <th>Rank</th>
                  <th>Username</th>
                  <th>Wins</th>
                  <th>Losses</th>
                  <th>Draws</th>
                </tr>
              </thead>
              <tbody>
                {!leaderboard || !Array.isArray(leaderboard) || leaderboard.length === 0 ? (
                  <tr>
                    <td colSpan="5" style={{ textAlign: 'center', color: '#999' }}>
                      No games played yet
                    </td>
                  </tr>
                ) : (
                  leaderboard.map((player, index) => (
                    <tr key={player.username || index}>
                      <td>{index + 1}</td>
                      <td><strong>{player.username || 'Unknown'}</strong></td>
                      <td>{player.wins || 0}</td>
                      <td>{player.losses || 0}</td>
                      <td>{player.draws || 0}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;

