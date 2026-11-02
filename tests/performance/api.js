import http from 'k6/http';
import { check, fail, sleep } from 'k6';

const profile = __ENV.PERF_PROFILE || 'smoke';
const profiles = { smoke: { vus: 3, duration: '30s' }, load: { vus: 20, duration: '120s' } };
if (!profiles[profile]) throw new Error('PERF_PROFILE deve ser smoke ou load');
const baseURL = 'http://api:8080';
const operations = ['login', 'abertura', 'consulta'];

export const options = {
  ...profiles[profile],
  setupTimeout: '60s',
  gracefulStop: '15s',
  systemTags: ['status', 'method', 'name', 'scenario', 'expected_response'],
  thresholds: {
    checks: ['rate==1'],
    http_req_failed: ['rate<0.01'],
    ...Object.fromEntries(operations.flatMap((operation) => [
      [`http_req_duration{operation:${operation},phase:load}`, ['p(95)<2000']],
      [`http_reqs{operation:${operation},phase:load}`, ['count>0']],
    ])),
  },
};

function json(response) {
  try { return response.json(); } catch (_) { return {}; }
}

function login(phase) {
  const response = http.post(`${baseURL}/autenticacao/login`, JSON.stringify({
    email: 'mecanico@oficina.local', senha: 'mecanico123',
  }), params('', 'login', phase));
  const body = json(response);
  const valid = check(response, {
    'login: HTTP 200 e token': (r) => r.status === 200 && typeof body.accessToken === 'string' && body.accessToken.length > 0,
  });
  return valid ? body.accessToken : '';
}

function params(token, operation, phase = 'load') {
  return {
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    tags: { name: operation, operation, phase },
    timeout: '10s',
  };
}

function cpf(number) {
  const digits = String(number).padStart(9, '0').split('').map(Number);
  for (let length = 9; length < 11; length++) {
    const remainder = digits.reduce((sum, digit, index) => sum + digit * (length + 1 - index), 0) * 10 % 11;
    digits.push(remainder === 10 ? 0 : remainder);
  }
  return digits.join('');
}

export function setup() {
  const token = login('setup');
  if (!token) fail('Falha ao autenticar a preparacao');
  const users = [];
  for (let index = 0; index < profiles[profile].vus; index++) {
    const customer = http.post(`${baseURL}/clientes`, JSON.stringify({
      nome: `Cliente desempenho ${index}`, documento: cpf(123450000 + index), tipoDocumento: 'CPF', telefone: '11999999999',
    }), params(token, 'preparar_cliente', 'setup'));
    const customerID = json(customer).id;
    if (!check(customer, { 'preparacao: cliente criado': (r) => r.status === 201 && !!customerID })) fail('Falha ao criar cliente de teste');
    const vehicle = http.post(`${baseURL}/clientes/${customerID}/veiculos`, JSON.stringify({
      placa: `KPF1A${String(index).padStart(2, '0')}`, marca: 'Fiat', modelo: 'Uno', ano: 2021,
    }), params(token, 'preparar_veiculo', 'setup'));
    const vehicleID = json(vehicle).id;
    if (!check(vehicle, { 'preparacao: veiculo criado': (r) => r.status === 201 && !!vehicleID })) fail('Falha ao criar veiculo de teste');
    users.push({ customerID, vehicleID });
  }
  return users;
}

export default function (users) {
  const token = login('load');
  if (!token) { sleep(0.5); return; }
  const user = users[__VU - 1];
  const response = http.post(`${baseURL}/ordens-servico`, JSON.stringify({
    clienteId: user.customerID,
    veiculoId: user.vehicleID,
    servicos: [{ servicoId: '40000000-0000-0000-0000-000000000001' }],
    pecas: [{ pecaId: '50000000-0000-0000-0000-000000000001', quantidade: 1 }],
  }), params(token, 'abertura'));
  const created = json(response);
  const valid = check(response, {
    'abertura: HTTP 201, identificador e RECEBIDA': (r) => r.status === 201 &&
      /^[0-9a-f-]{36}$/i.test(created.ordemServicoId || '') && created.status === 'RECEBIDA',
  });
  if (valid) {
    const consultation = http.get(`${baseURL}/ordens-servico/${created.ordemServicoId}`, params(token, 'consulta'));
    const order = json(consultation);
    check(consultation, {
      'consulta: HTTP 200 e mesma OS': (r) => r.status === 200 && order.ordemServicoId === created.ordemServicoId && order.statusOrdemServico === 'RECEBIDA',
      'consulta: orcamento com servico e peca': () => Array.isArray(order.orcamentos) && order.orcamentos.length === 1 &&
        order.orcamentos[0].itens.length === 2 && order.orcamentos[0].itens.some((item) => item.tipo === 'SERVICO') &&
        order.orcamentos[0].itens.some((item) => item.tipo === 'PECA'),
    });
  }
  sleep(0.5);
}

export function handleSummary(data) {
  const metrics = data.metrics;
  const failed = Object.values(metrics).some((metric) => Object.values(metric.thresholds || {}).some((threshold) => !threshold.ok));
  const lines = [
    '# Teste de desempenho da API', '',
    `Perfil: **${profile}** — ${profiles[profile].vus} usuarios virtuais por ${profiles[profile].duration}.`, '',
    `Resultado dos thresholds: **${failed ? 'REPROVADO' : 'APROVADO'}**.`, '',
    '| Operacao | Requisicoes | p95 (ms) | Limite p95 (ms) |', '| --- | ---: | ---: | ---: |',
    ...operations.map((operation) => {
      const duration = metrics[`http_req_duration{operation:${operation},phase:load}`];
      const requests = metrics[`http_reqs{operation:${operation},phase:load}`];
      return `| ${operation} | ${requests ? requests.values.count : 0} | ${duration ? duration.values['p(95)'].toFixed(2) : 'sem amostras'} | < 2000 |`;
    }), '',
    `Falhas HTTP: ${((metrics.http_req_failed ? metrics.http_req_failed.values.rate : 0) * 100).toFixed(2)}%.`,
    `Verificacoes funcionais: ${((metrics.checks ? metrics.checks.values.rate : 0) * 100).toFixed(2)}% aprovadas.`, '',
    'Ambiente temporario; estas medicoes nao representam capacidade de producao nem validam HPA.', '',
  ];
  return { '/results/summary.json': JSON.stringify(data, null, 2), '/results/summary.md': lines.join('\n'), stdout: lines.join('\n') };
}
