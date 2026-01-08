import http from 'k6/http';
import { check, sleep } from 'k6';

// Configuração dos estágios de carga
export const options = {
  stages: [
    { duration: '1s', target: 10 }, // Ramp-up: sobe para 10 usuários em 10s
    { duration: '3s', target: 10 }, // Platô: mantém 10 usuários por 30s
    { duration: '1s', target: 0 },  // Ramp-down: desce para 0 usuários
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% das requisições devem ser < 500ms
    http_req_failed: ['rate<0.01'],   // Menos de 1% de falhas permitidas
  },
};

export default function () {
  const url = 'http://localhost:8090/api/payments';
  
  // Payload para um pagamento autorizado (final do cartão 1 = ímpar)
  const payload = JSON.stringify({
    card_number: "1234567890123451",
    expiry_month: 12,
    expiry_year: 2030,
    currency: "USD",
    amount: 1500,
    cvv: "123"
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);

  // Verificações
  check(res, {
    'status is 200': (r) => r.status === 200,
    'payment is authorized': (r) => r.json('payment_status') === 'Authorized',
    'has payment id': (r) => r.json('id') !== '',
  });

  sleep(1);
}