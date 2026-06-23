package com.kyriosdata.assinador;

import com.kyriosdata.assinador.cli.CliRunner;
import com.kyriosdata.assinador.http.SignatureServer;

/**
 * Ponto de entrada do assinador.jar.
 *
 * <p>Modos de uso:
 * <ul>
 *   <li>{@code java -jar assinador.jar sign --content "..." [--token "..."]}</li>
 *   <li>{@code java -jar assinador.jar validate --content "..." --signature "..."}</li>
 *   <li>{@code java -jar assinador.jar server [--port 8080] [--timeout 30]}</li>
 * </ul>
 *
 * <p>Saída em stdout: JSON com campos {@code signature}, {@code valid}, {@code message}.
 * Erros vão para stderr; exit code 0 = sucesso, 1 = erro do usuário, 2 = erro do sistema.
 */
public class Main {

    public static void main(String[] args) {
        if (args.length == 0) {
            System.err.println("Uso: assinador.jar <comando> [opções]");
            System.err.println("Comandos: sign, validate, server");
            System.err.println("Use --help para detalhes.");
            System.exit(1);
        }

        String command = args[0];

        switch (command) {
            case "server" -> {
                SignatureServer server = new SignatureServer();
                server.run(args);
            }
            case "sign", "validate" -> {
                CliRunner runner = new CliRunner();
                runner.run(args);
            }
            case "--help", "-h", "help" -> printHelp();
            default -> {
                System.err.println("Comando desconhecido: " + command);
                System.err.println("Comandos disponíveis: sign, validate, server");
                System.exit(1);
            }
        }
    }

    private static void printHelp() {
        System.out.println("""
            assinador.jar — Simulador de Assinatura Digital

            COMANDOS:
              sign       Cria uma assinatura simulada
              validate   Valida uma assinatura simulada
              server     Inicia o servidor HTTP de assinatura

            USO:
              java -jar assinador.jar sign --content <conteudo> [--token <token>]
              java -jar assinador.jar validate --content <conteudo> --signature <assinatura>
              java -jar assinador.jar server [--port <porta>] [--timeout <minutos>]

            OPÇÕES COMUNS:
              --help, -h     Exibe esta ajuda

            SAÍDA:
              JSON em stdout: {"signature":"...","valid":true,"message":"..."}
              Erros em stderr. Exit code: 0=sucesso, 1=erro do usuário, 2=erro do sistema.
            """);
    }
}
