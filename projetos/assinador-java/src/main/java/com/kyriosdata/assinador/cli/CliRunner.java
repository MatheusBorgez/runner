package com.kyriosdata.assinador.cli;

import com.google.gson.Gson;
import com.kyriosdata.assinador.FakeSignatureService;
import com.kyriosdata.assinador.SignatureService;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

/**
 * Executa comandos {@code sign} e {@code validate} em modo CLI (invocação direta).
 *
 * <p>Lê parâmetros via flags {@code --content}, {@code --token}, {@code --signature}
 * e escreve o resultado como JSON em stdout.
 */
public class CliRunner {

    private final SignatureService service;
    private final Gson gson = new Gson();

    public CliRunner() {
        this.service = new FakeSignatureService();
    }

    public CliRunner(SignatureService service) {
        this.service = service;
    }

    public void run(String[] args) {
        String command = args[0];
        ArgParser parser = new ArgParser(args, 1);

        switch (command) {
            case "sign" -> handleSign(parser);
            case "validate" -> handleValidate(parser);
            default -> {
                System.err.println("Comando desconhecido: " + command);
                System.exit(1);
            }
        }
    }

    private void handleSign(ArgParser parser) {
        String content = parser.get("--content");
        String token = parser.get("--token");

        SignRequest request = new SignRequest();
        request.setContent(content);
        request.setToken(token);

        SignatureResponse response = service.sign(request);
        System.out.println(gson.toJson(response));

        if (!response.isValid()) {
            System.exit(1);
        }
    }

    private void handleValidate(ArgParser parser) {
        String content = parser.get("--content");
        String signature = parser.get("--signature");

        ValidateRequest request = new ValidateRequest();
        request.setContent(content);
        request.setSignature(signature);

        SignatureResponse response = service.validate(request);
        System.out.println(gson.toJson(response));

        if (!response.isValid()) {
            System.exit(1);
        }
    }
}
