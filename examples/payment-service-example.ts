/**
 * Example: Using the Charging Service API in a Payment Service
 *
 * This demonstrates how easy it is to integrate charging control
 * into your application without knowing OCPP protocol details.
 */

import { chargingService } from '../src/services/charging-service';
import {
  ChargePointNotFoundError,
  ChargePointOfflineError,
  ConnectorUnavailableError,
  SessionNotFoundError,
  getUserFriendlyErrorMessage
} from '../src/errors/custom-errors';

/**
 * Example Payment Service
 */
class PaymentService {
  /**
   * After user pays, start their charging session
   */
  async handlePaymentApproved(
    userId: string,
    chargePointId: string,
    connectorId: number,
    amountPaid: number
  ): Promise<void> {
    console.log(`\n💳 Payment approved for user ${userId}: $${amountPaid}`);
    console.log(`   Starting charging session on ${chargePointId} connector ${connectorId}...`);

    try {
      // Simple API - just turn on!
      const result = await chargingService.turnOn(chargePointId, connectorId, userId);

      console.log(`✅ ${result.message}`);
      console.log(`   Transaction ID: ${result.transactionId}`);
      console.log(`   User can now charge their vehicle!`);

      // Store transaction ID for later (in your database)
      await this.storeTransaction(userId, result.transactionId, amountPaid);
    } catch (error) {
      // Handle errors with user-friendly messages
      const userMessage = getUserFriendlyErrorMessage(error as Error);

      if (error instanceof ChargePointNotFoundError) {
        console.error(`❌ Error: ${userMessage}`);
        await this.refundPayment(userId, amountPaid);
        this.notifyUser(userId, 'Charging station not found. Your payment has been refunded.');
      } else if (error instanceof ChargePointOfflineError) {
        console.error(`❌ Error: ${userMessage}`);
        await this.refundPayment(userId, amountPaid);
        this.notifyUser(userId, 'Charging station is offline. Your payment has been refunded.');
      } else if (error instanceof ConnectorUnavailableError) {
        console.error(`❌ Error: ${userMessage}`);
        await this.refundPayment(userId, amountPaid);
        this.notifyUser(userId, 'Connector is in use. Your payment has been refunded.');
      } else {
        console.error(`❌ Unexpected error:`, error);
        await this.refundPayment(userId, amountPaid);
        this.notifyUser(userId, 'An error occurred. Your payment has been refunded.');
      }
    }
  }

  /**
   * Monitor energy consumption during charging
   */
  async checkChargingProgress(transactionId: number, chargePointId: string): Promise<void> {
    try {
      const consumption = await chargingService.getEnergyConsumption(chargePointId, transactionId);

      const energyKwh = consumption.currentEnergyWh / 1000;
      const durationMinutes = Math.floor(consumption.durationSeconds / 60);
      const cost = this.calculateCost(consumption.currentEnergyWh);

      console.log(`\n⚡ Charging Progress:`);
      console.log(`   Transaction: ${transactionId}`);
      console.log(`   User: ${consumption.userId}`);
      console.log(`   Energy: ${energyKwh.toFixed(2)} kWh`);
      console.log(`   Duration: ${durationMinutes} minutes`);
      console.log(`   Cost so far: $${cost.toFixed(2)}`);
      console.log(`   Status: ${consumption.status}`);
    } catch (error) {
      if (error instanceof SessionNotFoundError) {
        console.error(`❌ Session not found`);
      } else {
        console.error(`❌ Error checking progress:`, error);
      }
    }
  }

  /**
   * Stop charging session (user finished or balance depleted)
   */
  async stopChargingSession(
    transactionId: number,
    chargePointId: string,
    reason: string
  ): Promise<void> {
    console.log(`\n🛑 Stopping charging session ${transactionId}...`);
    console.log(`   Reason: ${reason}`);

    try {
      // Simple API - just turn off!
      const result = await chargingService.turnOff(chargePointId, transactionId);

      const energyKwh = result.energyConsumed / 1000;
      const durationMinutes = Math.floor(result.duration / 60);
      const finalCost = this.calculateCost(result.energyConsumed);

      console.log(`✅ ${result.message}`);
      console.log(`\n📊 Session Summary:`);
      console.log(`   Total Energy: ${energyKwh.toFixed(2)} kWh`);
      console.log(`   Duration: ${durationMinutes} minutes`);
      console.log(`   Final Cost: $${finalCost.toFixed(2)}`);

      // Process final payment
      await this.processFinalPayment(transactionId, finalCost);
    } catch (error) {
      const userMessage = getUserFriendlyErrorMessage(error as Error);
      console.error(`❌ Error stopping session: ${userMessage}`);
    }
  }

  // Mock helper methods

  private async storeTransaction(
    userId: string,
    transactionId: number,
    amountPaid: number
  ): Promise<void> {
    // In real app: store in your database
    console.log(`   💾 Stored transaction ${transactionId} in database`);
  }

  private async refundPayment(userId: string, amount: number): Promise<void> {
    // In real app: process refund
    console.log(`   💰 Refunded $${amount} to user ${userId}`);
  }

  private notifyUser(userId: string, message: string): void {
    // In real app: send push notification, SMS, etc.
    console.log(`   📧 Notified user ${userId}: ${message}`);
  }

  private calculateCost(energyWh: number): number {
    // $0.30 per kWh
    const pricePerKwh = 0.3;
    const energyKwh = energyWh / 1000;
    return energyKwh * pricePerKwh;
  }

  private async processFinalPayment(transactionId: number, amount: number): Promise<void> {
    // In real app: charge credit card, update balance, etc.
    console.log(`   💳 Processed final payment: $${amount.toFixed(2)}`);
  }
}

/**
 * Demo scenario
 */
async function demo() {
  const paymentService = new PaymentService();

  console.log('═══════════════════════════════════════════════════════');
  console.log('         Payment Service Integration Demo');
  console.log('═══════════════════════════════════════════════════════');

  // Scenario 1: User pays and starts charging
  await paymentService.handlePaymentApproved('user-123', 'CP001', 1, 25.0);

  // Wait a bit (in real app, this would be periodic checks)
  await new Promise((resolve) => setTimeout(resolve, 2000));

  // Scenario 2: Check charging progress
  // Note: In real usage with actual charge points, there would be a real transaction
  // await paymentService.checkChargingProgress(1, 'CP001');

  // Scenario 3: Stop charging
  // await paymentService.stopChargingSession(1, 'CP001', 'User requested stop');

  console.log('\n═══════════════════════════════════════════════════════');
  console.log('         Demo Complete');
  console.log('═══════════════════════════════════════════════════════\n');
}

// Run demo if executed directly
if (require.main === module) {
  // Note: This requires the OCPP server to be running
  console.log('⚠️  Make sure the OCPP server is running (npm run dev)');
  console.log('    And charge points are connected before running this demo\n');

  setTimeout(() => {
    demo().catch(console.error);
  }, 1000);
}

export { PaymentService };
