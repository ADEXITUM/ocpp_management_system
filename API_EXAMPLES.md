# REST API Examples

REST API для управления зарядными сессиями.

## Запуск сервера

```bash
go run cmd/server/main.go
```

Сервер запустит:
- **OCPP WebSocket** на порту `9000` (для зарядных станций)
- **REST API** на порту `8080` (для вашего приложения)

## Эндпоинты

### 1. Начать зарядную сессию (с оплатой)

**POST** `http://localhost:8080/sessions/start`

```bash
curl -X POST http://localhost:8080/sessions/start \
  -H "Content-Type: application/json" \
  -d '{
    "chargePointId": "CP001",
    "connectorId": 1,
    "userId": "user-123",
    "amountPaid": 25.00
  }'
```

**Ответ (успех):**
```json
{
  "success": true,
  "message": "Charging session started successfully",
  "transactionId": 1,
  "chargePointId": "CP001",
  "connectorId": 1,
  "userId": "user-123",
  "paymentStatus": "ОПЛАТА ПРОШЛА УСПЕШНО ✓"
}
```

**Ответ (ошибка - станция оффлайн):**
```json
{
  "success": false,
  "error": "Charge point 'CP001' is not connected",
  "errorCode": "CHARGE_POINT_OFFLINE",
  "userMessage": "This charging station is currently offline. Please try another station or check back later."
}
```

### 2. Получить текущее потребление энергии

**GET** `http://localhost:8080/sessions/{transactionId}/energy?chargePointId=CP001`

```bash
curl "http://localhost:8080/sessions/1/energy?chargePointId=CP001"
```

**Ответ:**
```json
{
  "success": true,
  "transactionId": 1,
  "chargePointId": "CP001",
  "connectorId": 1,
  "userId": "user-123",
  "energyWh": 2500.5,
  "energyKwh": 2.5005,
  "durationSeconds": 600,
  "durationMinutes": 10,
  "status": "active",
  "costEstimate": 0.75
}
```

### 3. Остановить зарядную сессию

**POST** `http://localhost:8080/sessions/{transactionId}/stop?chargePointId=CP001`

```bash
curl -X POST "http://localhost:8080/sessions/1/stop?chargePointId=CP001"
```

**Ответ:**
```json
{
  "success": true,
  "message": "Charging session stopped successfully",
  "transactionId": 1,
  "chargePointId": "CP001",
  "energyConsumed": 5000,
  "energyKwh": 5.0,
  "durationSeconds": 1200,
  "durationMinutes": 20,
  "finalCost": 1.5
}
```

### 4. Проверить здоровье сервиса

**GET** `http://localhost:8080/health`

```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "status": "ok",
  "service": "ocpp-management-system"
}
```

## Полный пример использования

### Шаг 1: Запустить сервер

```bash
# Terminal 1
go run cmd/server/main.go
```

### Шаг 2: Подключить эмулятор зарядной станции

```bash
# Terminal 2
cd ocpp-virtual-charge-point
WS_URL=ws://localhost:9000/CP001 npx tsx index_16.ts
```

### Шаг 3: Использовать API

```bash
# Terminal 3

# 1. Начать сессию (с оплатой)
curl -X POST http://localhost:8080/sessions/start \
  -H "Content-Type: application/json" \
  -d '{
    "chargePointId": "CP001",
    "connectorId": 1,
    "userId": "user-123",
    "amountPaid": 25.00
  }'

# Сохраните transactionId из ответа (например, 1)

# 2. Проверить энергию (подождите несколько секунд)
curl "http://localhost:8080/sessions/1/energy?chargePointId=CP001"

# 3. Остановить сессию
curl -X POST "http://localhost:8080/sessions/1/stop?chargePointId=CP001"
```

## Коды ошибок

| Код | HTTP Status | Описание |
|-----|-------------|----------|
| `CHARGE_POINT_NOT_FOUND` | 404 | Зарядная станция не найдена в базе |
| `CHARGE_POINT_OFFLINE` | 409 | Зарядная станция не подключена |
| `CONNECTOR_NOT_FOUND` | 404 | Коннектор не существует |
| `CONNECTOR_UNAVAILABLE` | 409 | Коннектор занят или неисправен |
| `SESSION_NOT_FOUND` | 404 | Сессия не найдена |
| `SESSION_ALREADY_ACTIVE` | 409 | На коннекторе уже активна сессия |
| `OPERATION_TIMEOUT` | 408 | Операция превысила тайм-аут |
| `REMOTE_OPERATION_FAILED` | 400 | Зарядная станция отклонила команду |

## Переменные окружения

```bash
# OCPP WebSocket порт (по умолчанию 9000)
export OCPP_PORT=9000

# REST API порт (по умолчанию 8080)
export API_PORT=8080
```

## Тестирование с Postman

Импортируйте следующую коллекцию:

```json
{
  "info": {
    "name": "OCPP Management System",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Start Session",
      "request": {
        "method": "POST",
        "header": [{"key": "Content-Type", "value": "application/json"}],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"chargePointId\": \"CP001\",\n  \"connectorId\": 1,\n  \"userId\": \"user-123\",\n  \"amountPaid\": 25.00\n}"
        },
        "url": "http://localhost:8080/sessions/start"
      }
    },
    {
      "name": "Get Energy",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:8080/sessions/1/energy?chargePointId=CP001",
          "query": [{"key": "chargePointId", "value": "CP001"}]
        }
      }
    },
    {
      "name": "Stop Session",
      "request": {
        "method": "POST",
        "url": {
          "raw": "http://localhost:8080/sessions/1/stop?chargePointId=CP001",
          "query": [{"key": "chargePointId", "value": "CP001"}]
        }
      }
    }
  ]
}
```

## JavaScript/TypeScript Example

```typescript
// Начать сессию
async function startChargingSession(userId: string, amount: number) {
  const response = await fetch('http://localhost:8080/sessions/start', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      chargePointId: 'CP001',
      connectorId: 1,
      userId: userId,
      amountPaid: amount
    })
  });

  const data = await response.json();

  if (data.success) {
    console.log('Session started:', data.transactionId);
    console.log(data.paymentStatus); // "ОПЛАТА ПРОШЛА УСПЕШНО ✓"
    return data.transactionId;
  } else {
    console.error('Error:', data.userMessage);
    throw new Error(data.userMessage);
  }
}

// Получить энергию
async function getEnergy(transactionId: number) {
  const response = await fetch(
    `http://localhost:8080/sessions/${transactionId}/energy?chargePointId=CP001`
  );

  const data = await response.json();
  console.log(`Energy: ${data.energyKwh} kWh`);
  console.log(`Cost: $${data.costEstimate}`);
  return data;
}

// Остановить сессию
async function stopSession(transactionId: number) {
  const response = await fetch(
    `http://localhost:8080/sessions/${transactionId}/stop?chargePointId=CP001`,
    { method: 'POST' }
  );

  const data = await response.json();
  console.log(`Final cost: $${data.finalCost}`);
  return data;
}

// Использование
(async () => {
  try {
    const txId = await startChargingSession('user-123', 25.00);

    // Проверить энергию через 30 секунд
    setTimeout(async () => {
      await getEnergy(txId);
    }, 30000);

    // Остановить через 1 минуту
    setTimeout(async () => {
      await stopSession(txId);
    }, 60000);
  } catch (error) {
    console.error('Error:', error);
  }
})();
```

## Python Example

```python
import requests
import time

BASE_URL = "http://localhost:8080"

# Начать сессию
def start_session(user_id, amount):
    response = requests.post(f"{BASE_URL}/sessions/start", json={
        "chargePointId": "CP001",
        "connectorId": 1,
        "userId": user_id,
        "amountPaid": amount
    })

    data = response.json()
    if data["success"]:
        print(f"Session started: {data['transactionId']}")
        print(data["paymentStatus"])
        return data["transactionId"]
    else:
        raise Exception(data["userMessage"])

# Получить энергию
def get_energy(transaction_id):
    response = requests.get(
        f"{BASE_URL}/sessions/{transaction_id}/energy",
        params={"chargePointId": "CP001"}
    )

    data = response.json()
    print(f"Energy: {data['energyKwh']} kWh")
    print(f"Cost: ${data['costEstimate']}")
    return data

# Остановить сессию
def stop_session(transaction_id):
    response = requests.post(
        f"{BASE_URL}/sessions/{transaction_id}/stop",
        params={"chargePointId": "CP001"}
    )

    data = response.json()
    print(f"Final cost: ${data['finalCost']}")
    return data

# Использование
if __name__ == "__main__":
    try:
        tx_id = start_session("user-123", 25.00)

        time.sleep(30)  # Подождать 30 секунд
        get_energy(tx_id)

        time.sleep(30)  # Подождать еще 30 секунд
        stop_session(tx_id)
    except Exception as e:
        print(f"Error: {e}")
```
