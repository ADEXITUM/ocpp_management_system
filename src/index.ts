/**
 * OCPP Management System - Main Entry Point
 */

import { OCPPServer } from './ocpp/ocpp-server';
import { chargingService } from './services/charging-service';
import { database } from './database/mock-database';
import { connectionManager } from './ocpp/connection-manager';

const PORT = process.env.OCPP_PORT ? parseInt(process.env.OCPP_PORT) : 9000;

async function main() {
  console.log('═══════════════════════════════════════════════════════');
  console.log('         OCPP Management System v1.0.0');
  console.log('         OCPP 1.6J WebSocket Server');
  console.log('═══════════════════════════════════════════════════════\n');

  // Display configured charge points
  const chargePoints = database.getAllChargePoints();
  console.log(`📊 Configured Charge Points: ${chargePoints.length}`);
  chargePoints.forEach((cp) => {
    console.log(`   - ${cp.id}: ${cp.name} (${cp.vendor} ${cp.model})`);
    console.log(`     Connectors: ${cp.numberOfConnectors}, Status: ${cp.status}`);
  });
  console.log('');

  // Start OCPP server
  const server = new OCPPServer(PORT);
  await server.start();

  console.log('');
  console.log('✅ System ready to accept charge point connections');
  console.log('   Charge points should connect to: ws://<server-ip>:9000/<charge-point-id>');
  console.log('   Example: ws://localhost:9000/CP001\n');

  // Display service API info
  console.log('📡 Service API Available:');
  console.log('   - chargingService.turnOn(chargePointId, connectorId, userId)');
  console.log('   - chargingService.turnOff(chargePointId, transactionId)');
  console.log('   - chargingService.getEnergyConsumption(chargePointId, transactionId)');
  console.log('');

  // Set up periodic status display
  setInterval(() => {
    displayStatus();
  }, 60000); // Every minute

  // Handle shutdown
  process.on('SIGINT', async () => {
    console.log('\n\n🛑 Shutting down OCPP Management System...');
    await server.stop();
    process.exit(0);
  });

  process.on('SIGTERM', async () => {
    console.log('\n\n🛑 Shutting down OCPP Management System...');
    await server.stop();
    process.exit(0);
  });
}

function displayStatus() {
  const dbStats = database.getStats();
  const connStats = connectionManager.getStats();

  console.log('\n─────────────────────────────────────────────────────');
  console.log(`📊 System Status - ${new Date().toLocaleString()}`);
  console.log(`   Charge Points: ${dbStats.onlineChargePoints}/${dbStats.totalChargePoints} online`);
  console.log(`   Active Sessions: ${dbStats.activeSessions}`);
  console.log(`   WebSocket Connections: ${connStats.totalConnections}`);
  console.log('─────────────────────────────────────────────────────\n');
}

// Export for use in other modules
export { chargingService, database, connectionManager };

// Start the server
if (require.main === module) {
  main().catch((error) => {
    console.error('❌ Failed to start OCPP Management System:', error);
    process.exit(1);
  });
}
