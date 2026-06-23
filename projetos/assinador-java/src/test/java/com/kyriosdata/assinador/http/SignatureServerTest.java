package com.kyriosdata.assinador.http;

import com.google.gson.Gson;
import com.kyriosdata.assinador.domain.SignatureResponse;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Base64;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Testes de integração do servidor HTTP.
 * Inicia um servidor em porta dinâmica e valida os endpoints.
 */
class SignatureServerTest {

    private static final int TEST_PORT = 19080;

    private SignatureServer server;
    private Thread serverThread;
    private final HttpClient http = HttpClient.newHttpClient();
    private final Gson gson = new Gson();

    @BeforeEach
    void startServer() throws Exception {
        server = new SignatureServer();
        serverThread = new Thread(() -> {
            try {
                server.start(TEST_PORT, 0);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt(); // restaura flag de interrupção
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }, "test-server");
        serverThread.setDaemon(true);
        serverThread.start();
        waitForServer(TEST_PORT, 3_000);
    }

    @AfterEach
    void stopServer() {
        server.stop();
        serverThread.interrupt();
    }

    // ─── /health ─────────────────────────────────────────────────────────────

    @Test
    void healthDeveRetornar200() throws Exception {
        HttpResponse<String> response = get("/health");
        assertEquals(200, response.statusCode());
        assertTrue(response.body().contains("ok"));
    }

    // ─── /sign ───────────────────────────────────────────────────────────────

    @Test
    void signDeveRetornar200ParaContentBase64Valido() throws Exception {
        String body = "{\"content\":\"" + b64("conteudo válido") + "\"}";
        HttpResponse<String> response = post("/sign", body);

        assertEquals(200, response.statusCode());
        SignatureResponse resp = gson.fromJson(response.body(), SignatureResponse.class);
        assertTrue(resp.isValid());
        assertNotNull(resp.getSignature());
    }

    @Test
    void signDeveRetornar422ParaContentAusente() throws Exception {
        HttpResponse<String> response = post("/sign", "{}");

        assertEquals(422, response.statusCode());
        SignatureResponse resp = gson.fromJson(response.body(), SignatureResponse.class);
        assertFalse(resp.isValid());
        assertTrue(resp.getMessage().contains("content"));
    }

    @Test
    void signDeveRetornar405ParaGetRequest() throws Exception {
        HttpResponse<String> response = get("/sign");
        assertEquals(405, response.statusCode());
    }

    // ─── /validate ───────────────────────────────────────────────────────────

    @Test
    void validateDeveRetornar200ParaAssinaturaCorreta() throws Exception {
        String body = "{\"content\":\"" + b64("x") + "\",\"signature\":\"MOCKED_SIGNATURE_BASE64_==\"}";
        HttpResponse<String> response = post("/validate", body);

        assertEquals(200, response.statusCode());
        SignatureResponse resp = gson.fromJson(response.body(), SignatureResponse.class);
        assertTrue(resp.isValid());
    }

    @Test
    void validateDeveRetornar422ParaAssinaturaInvalida() throws Exception {
        String body = "{\"content\":\"" + b64("x") + "\",\"signature\":\"ASSINATURA_ERRADA_==\"}";
        HttpResponse<String> response = post("/validate", body);

        assertEquals(422, response.statusCode());
        SignatureResponse resp = gson.fromJson(response.body(), SignatureResponse.class);
        assertFalse(resp.isValid());
    }

    @Test
    void validateDeveRetornar422QuandoSignatureAusente() throws Exception {
        String body = "{\"content\":\"" + b64("x") + "\"}";
        HttpResponse<String> response = post("/validate", body);

        assertEquals(422, response.statusCode());
    }

    // ─── /shutdown ────────────────────────────────────────────────────────────

    @Test
    void shutdownDeveRetornar200EEncerrarServidor() throws Exception {
        HttpResponse<String> response = post("/shutdown", "");
        assertEquals(200, response.statusCode());
        assertTrue(response.body().contains("shutdown"));
    }

    // ─── helpers ─────────────────────────────────────────────────────────────

    private HttpResponse<String> get(String path) throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:" + TEST_PORT + path))
            .GET()
            .build();
        return http.send(request, HttpResponse.BodyHandlers.ofString());
    }

    private HttpResponse<String> post(String path, String body) throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create("http://localhost:" + TEST_PORT + path))
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build();
        return http.send(request, HttpResponse.BodyHandlers.ofString());
    }

    private String b64(String text) {
        return Base64.getEncoder().encodeToString(text.getBytes());
    }

    private void waitForServer(int port, long timeoutMs) throws Exception {
        long deadline = System.currentTimeMillis() + timeoutMs;
        while (System.currentTimeMillis() < deadline) {
            try {
                HttpRequest req = HttpRequest.newBuilder()
                    .uri(URI.create("http://localhost:" + port + "/health"))
                    .GET().build();
                http.send(req, HttpResponse.BodyHandlers.ofString());
                return;
            } catch (IOException e) {
                Thread.sleep(100);
            }
        }
        throw new RuntimeException("Servidor não iniciou em " + timeoutMs + "ms");
    }
}
