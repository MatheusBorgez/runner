package com.kyriosdata.assinador;

import com.kyriosdata.assinador.cli.CliRunner;
import com.kyriosdata.assinador.http.SignatureServer;

public class Main {

    public static void main(String[] args) {
        if (args.length == 0) {
            System.err.println("Uso: assinador.jar <comando> [opções]");
            System.err.println("Comandos: sign, validate, server");
            System.err.println("Use --help para detalhes.");
            System.exit(1);
        }

        switch (args[0]) {
            case "server" -> new SignatureServer().run(args);
            case "sign", "validate" -> new CliRunner().run(args);
            case "--help", "-h", "help" -> printHelp();
            default -> {
                System.err.println("Comando desconhecido: " + args[0]);
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

            SAÍDA:
              JSON em stdout: {"signature":"...","valid":true,"message":"..."}
              Erros em stderr. Exit code: 0=sucesso, 1=erro do usuário, 2=erro do sistema.
            """);
    }
}
