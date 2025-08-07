const WebSocket = require('ws');

// Test legacy WebSocket endpoint
console.log('Testing legacy WebSocket endpoint...');
const legacyWs = new WebSocket('ws://localhost:8098/ws/connect');

legacyWs.on('open', function open() {
  console.log('✓ Legacy WebSocket connected successfully');
  
  // Test sending a message
  legacyWs.send(JSON.stringify({
    type: 'subscribe',
    channel: 'project:1'
  }));
});

legacyWs.on('message', function message(data) {
  console.log('Legacy WebSocket received:', data.toString());
});

legacyWs.on('error', function error(err) {
  console.log('✗ Legacy WebSocket error:', err.message);
});

legacyWs.on('close', function close(code, reason) {
  console.log('Legacy WebSocket closed:', code, reason.toString());
});

// Test enhanced service endpoint detection
setTimeout(() => {
  console.log('\nTesting enhanced service endpoint detection...');
  
  const testWs = new WebSocket('ws://localhost:8098/ws/enhanced');
  
  testWs.on('open', function open() {
    console.log('✓ Enhanced WebSocket endpoint responding');
  });
  
  testWs.on('error', function error(err) {
    console.log('✗ Enhanced WebSocket endpoint error:', err.message);
  });
  
  testWs.on('close', function close() {
    console.log('Enhanced WebSocket endpoint closed');
    process.exit(0);
  });
  
  setTimeout(() => {
    testWs.close();
    legacyWs.close();
  }, 2000);
  
}, 1000);