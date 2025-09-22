import React, { useState, useRef, useEffect } from 'react';
import './MIAConsole.css';

const MIAConsole = () => {
  const [command, setCommand] = useState('');
  const [output, setOutput] = useState([]);
  const [commandHistory, setCommandHistory] = useState([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const [selectedFile, setSelectedFile] = useState(null);
  const outputRef = useRef(null);
  const fileInputRef = useRef(null);

  // Auto-scroll to bottom when new output is added
  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [output]);

  // Initialize empty output
  useEffect(() => {
    setOutput([]);
  }, []);

  const processMultipleCommands = async (text) => {
    const lines = text.split('\n');

    for (const line of lines) {
      const trimmedLine = line.trim();

      // Skip empty lines
      if (!trimmedLine) {
        continue;
      }

      // Handle comments - show them but don't execute
      if (trimmedLine.startsWith('#')) {
        const commentOutput = {
          type: 'comment',
          content: trimmedLine,
          timestamp: new Date().toLocaleTimeString()
        };
        setOutput(prev => [...prev, commentOutput]);
        continue;
      }

      // Execute actual commands
      await executeCommand(trimmedLine, false); // false = don't add to history for batch execution

      // Small delay between commands for better UX
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  };

  const executeCommand = async (cmd, addToHistory = true) => {
    if (!cmd.trim()) return;

    // Add command to history only for individual executions
    if (addToHistory) {
      setCommandHistory(prev => [...prev, cmd]);
      setHistoryIndex(-1);
    }

    // Add command to output
    const newOutput = {
      type: 'command',
      content: `MIA> ${cmd}`,
      timestamp: new Date().toLocaleTimeString()
    };

    setOutput(prev => [...prev, newOutput]);

    // Handle special frontend commands
    if (cmd.toLowerCase() === 'clear') {
      setOutput([]);
      return;
    }

    if (cmd.toLowerCase() === 'help') {
      const helpOutput = {
        type: 'success',
        content: `Comandos disponibles:
• mkdisk -size=[tamaño] -path=[ruta] -unit=[unit]
• rmdisk -path=[ruta]
• fdisk -size=[tamaño] -path=[ruta] -name=[nombre] -type=[tipo] -fit=[ajuste] -unit=[unit]
• mount -path=[ruta] -name=[nombre]
• mounted
• mkfs -id=[id] -type=[tipo] -fs=[sistema]
• cat -file=[archivo] -id=[id]
• showdisk -path=[ruta]
• clear - Limpiar consola
• exit - Salir del programa`,
        timestamp: new Date().toLocaleTimeString()
      };
      setOutput(prev => [...prev, helpOutput]);
      return;
    }

    // Send command to backend
    try {
      const response = await fetch('http://localhost:8080/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command: cmd }),
      });

      const result = await response.json();

      const resultOutput = {
        type: result.error ? 'error' : 'success',
        content: result.error || result.output || 'Comando ejecutado correctamente',
        timestamp: new Date().toLocaleTimeString()
      };

      setOutput(prev => [...prev, resultOutput]);
    } catch (error) {
      const errorOutput = {
        type: 'error',
        content: `Error de conexión: ${error.message}\nAsegúrate de que el backend esté ejecutándose en el puerto 8080`,
        timestamp: new Date().toLocaleTimeString()
      };
      setOutput(prev => [...prev, errorOutput]);
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    // Check if command contains multiple lines (from paste)
    if (command.includes('\n')) {
      processMultipleCommands(command);
    } else {
      executeCommand(command);
    }

    setCommand('');
  };

  const handlePaste = (e) => {
    // Allow default paste behavior - the multiline handling will be done in handleSubmit
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSubmit(e);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (commandHistory.length > 0) {
        const newIndex = historyIndex === -1 ? commandHistory.length - 1 : Math.max(0, historyIndex - 1);
        setHistoryIndex(newIndex);
        setCommand(commandHistory[newIndex]);
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex >= 0) {
        const newIndex = historyIndex + 1;
        if (newIndex >= commandHistory.length) {
          setHistoryIndex(-1);
          setCommand('');
        } else {
          setHistoryIndex(newIndex);
          setCommand(commandHistory[newIndex]);
        }
      }
    }
  };

  const handleFileSelect = (e) => {
    const file = e.target.files[0];
    if (file) {
      const fileExtension = file.name.split('.').pop().toLowerCase();
      if (fileExtension === 'txt' || fileExtension === 'mia') {
        setSelectedFile(file);
      } else {
        const errorOutput = {
          type: 'error',
          content: 'Solo se permiten archivos .txt y .mia',
          timestamp: new Date().toLocaleTimeString()
        };
        setOutput(prev => [...prev, errorOutput]);
        e.target.value = '';
      }
    }
  };

  const handleFileExecute = async () => {
    if (!selectedFile) return;

    try {
      const fileContent = await selectedFile.text();

      const fileOutput = {
        type: 'success',
        content: `Ejecutando archivo: ${selectedFile.name}`,
        timestamp: new Date().toLocaleTimeString()
      };
      setOutput(prev => [...prev, fileOutput]);

      await processMultipleCommands(fileContent);

      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    } catch (error) {
      const errorOutput = {
        type: 'error',
        content: `Error al leer el archivo: ${error.message}`,
        timestamp: new Date().toLocaleTimeString()
      };
      setOutput(prev => [...prev, errorOutput]);
    }
  };

  return (
    <div className="mia-console">
      <div className="console-container">
        {/* Cuadro de entrada de comandos */}
        <div className="input-section">
          <div className="section-header">
            <h3>Entrada de Comandos</h3>
          </div>

          {/* File upload section */}
          <div className="file-upload-section">
            <div className="file-upload-line">
              <input
                type="file"
                ref={fileInputRef}
                onChange={handleFileSelect}
                accept=".txt,.mia"
                className="file-input"
              />
              <button
                type="button"
                onClick={handleFileExecute}
                disabled={!selectedFile}
                className="execute-btn"
              >
                Ejecutar Archivo
              </button>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="console-input">
            <div className="input-line">
              <span className="prompt">MIA&gt;</span>
              <textarea
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                onKeyDown={handleKeyDown}
                onPaste={handlePaste}
                className="command-input"
                placeholder="Ingresa un comando o pega múltiples líneas..."
                autoFocus
                rows={command.includes('\n') ? Math.min(command.split('\n').length, 10) : 1}
              />
              <button type="submit" className="execute-btn">
                Ejecutar
              </button>
            </div>
          </form>
        </div>

        {/* Cuadro de salida */}
        <div className="output-section">
          <div className="section-header">
            <h3>Salida del Sistema</h3>
          </div>
          <div className="console-output" ref={outputRef}>
            {output.map((item, index) => (
              <div key={index} className={`output-line ${item.type}`}>
                <span className="timestamp">[{item.timestamp}]</span>
                <pre className="content">{item.content}</pre>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default MIAConsole;