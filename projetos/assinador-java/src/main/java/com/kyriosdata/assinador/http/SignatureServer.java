package com.kyriosdata.assinador.http;

import com.google.gson.Gson;
import com.kyriosdata.assinador.FakeSignatureService;
import com.kyriosdata.assinador.SignatureService;
import com.kyriosdata.assinador.cli.ArgParser;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;
import java.util.logging.Logger;

/**
 * Servidor HTTP do assinador.jar.
 *
 * <p>Expõe os endpoints:
 * <ul>
 *   <li>{@code POST /sign}      — cria assinatura simulada</li>
 *   <li>{@code POST /validate}  — valida assinatura simulada</li>
 *   <li>{@code GET  /health}    — verifica saúde do servidor</li>
 *   <li>{@code POST /shutdown}  — encerra o servidor de forma limpa</li>
 * </ul>
 *
 * <p>Porta padrão: 8080. Configurável via {@code --port}.
 * Auto-shutdown por inatividade configurável via {@code --timeout} (minutos).
 */
public class SignatureServer {

    static final int DEFAULT_PORT = 8080;
    static final int DEFAULT_TIMEOUT_MINUTES = 0; // 0 = sem timeout

    private static final Logger LOG = Logger.getLogger(SignatureServer.class.getName());

    private final SignatureService service;
    private final Gson gson = new Gson();
    private final AtomicLong lastActivityMs = new AtomicLong(System.currentTimeMillis());

    private HttpServer server;
    private ScheduledExecutorService scheduler;

    public SignatureServer() {
        this.service = new FakeSignatureService();
    }

    public SignatureServer(SignatureService service) {
        this.service = service;
    }

    public void run(String[] args) {
        ArgParser parser = new ArgParser(args, 1);
        int port = parser.getInt("--port", DEFAULT_PORT);
        int timeoutMinutes = parser.getInt("--timeout", DEFAULT_TIMEOUT_MINUTES);

        try {
            start(port, timeoutMinutes);
        } catch (IOException e) {
            System.err.println("Erro ao iniciar servidor na porta " + port + ": " + e.getMessage());
            System.err.println("Verifique se a porta está disponível e tente novamente.");
            System.exit(2);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    public void start(int port, int timeoutMinutes) throws IOException, InterruptedException {
        server = HttpServer.create(new InetSocketAddress(port), 0);
        server.createContext("/sign", this::handleSign);
        server.createContext("/validate", this::handleValidate);
        server.createContext("/health", this::handleHealth);
        server.createContext("/shutdown", this::handleShutdown);
        server.setExecutor(Executors.newVirtualThreadPerTaskExecutor());
        server.start();

        System.out.println("{\"status\":\"started\",\"port\":" + port + "}");
        LOG.info("Servidor iniciado na porta " + port);

        if (timeoutMinutes > 0) {
            scheduleInactivityShutdown(timeoutMinutes);
        }

        // bloqueia thread principal
        Thread.currentThread().join();
    }

    private void scheduleInactivityShutdown(int timeoutMinutes) {
        scheduler = Executors.newSingleThreadScheduledExecutor(r -> {
            Thread t = new Thread(r, "inactivity-watchdog");
            t.setDaemon(true);
            return t;
        });

        long checkIntervalMs = Math.max(10_000, timeoutMinutes * 60_000L / 10);

        timeoutTask = scheduler.scheduleAtFixedRate(() -> {
            long idleMs = System.currentTimeMillis() - lastActivityMs.get();
            if (idleMs >= timeoutMinutes * 60_000L) {
                LOG.info("Timeout de inatividade atingido (" + timeoutMinutes + " min). Encerrando.");
                System.out.println("{\"status\":\"shutdown\",\"reason\":\"inactivity\"}");
                stop();
            }
        }, checkIntervalMs, checkIntervalMs, TimeUnit.MILLISECONDS);
    }

    public void stop() {
        if (server != null) {
            server.stop(1);
        }
        if (scheduler != null) {
            scheduler.shutdownNow();
        }
    }

    private void handleSign(HttpExchange exchange) throws IOException {
        recordActivity();
        if (!"POST".equalsIgnoreCase(exchange.getRequestMethod())) {
            sendResponse(exchange, 405, "{\"valid\":false,\"message\":\"Método não permitido; use POST\"}");
            return;
        }
        String body = readBody(exchange);
        SignRequest request = gson.fromJson(body, SignRequest.class);
        SignatureResponse response = service.sign(request);
        int status = response.isValid() ? 200 : 422;
        sendResponse(exchange, status, gson.toJson(response));
    }

    private void handleValidate(HttpExchange exchange) throws IOException {
        recordActivity();
        if (!"POST".equalsIgnoreCase(exchange.getRequestMethod())) {
            sendResponse(exchange, 405, "{\"valid\":false,\"message\":\"Método não permitido; use POST\"}");
            return;
        }
        String body = readBody(exchange);
        ValidateRequest request = gson.fromJson(body, ValidateRequest.class);
        SignatureResponse response = service.validate(request);
        int status = response.isValid() ? 200 : 422;
        sendResponse(exchange, status, gson.toJson(response));
    }

    private void handleHealth(HttpExchange exchange) throws IOException {
        recordActivity();
        sendResponse(exchange, 200, "{\"status\":\"ok\"}");
    }

    private void handleShutdown(HttpExchange exchange) throws IOException {
        if (!"POST".equalsIgnoreCase(exchange.getRequestMethod())) {
            sendResponse(exchange, 405, "{\"status\":\"error\",\"message\":\"Método não permitido; use POST\"}");
            return;
        }
        sendResponse(exchange, 200, "{\"status\":\"shutdown\"}");
        LOG.info("Shutdown solicitado via endpoint /shutdown.");
        new Thread(() -> {
            try { Thread.sleep(200); } catch (InterruptedException ignored) { Thread.currentThread().interrupt(); }
            stop();
        }, "shutdown-thread").start();
    }

    private void recordActivity() {
        // Apenas atualiza o timestamp; o watchdog verifica periodicamente
        lastActivityMs.set(System.currentTimeMillis());
    }

    private String readBody(HttpExchange exchange) throws IOException {
        try (InputStream is = exchange.getRequestBody()) {
            return new String(is.readAllBytes(), StandardCharsets.UTF_8);
        }
    }

    private void sendResponse(HttpExchange exchange, int status, String json) throws IOException {
        byte[] bytes = json.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().set("Content-Type", "application/json; charset=UTF-8");
        exchange.sendResponseHeaders(status, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }
}
