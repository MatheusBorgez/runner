package com.kyriosdata.assinador.cli;

import java.util.HashMap;
import java.util.Map;

public class ArgParser {

    private final Map<String, String> values = new HashMap<>();

    public ArgParser(String[] args, int startIndex) {
        for (int i = startIndex; i < args.length - 1; i++) {
            if (args[i].startsWith("--")) {
                values.put(args[i], args[i + 1]);
                i++;
            }
        }
    }

    public String get(String flag) {
        return values.get(flag);
    }

    public String getOrDefault(String flag, String defaultValue) {
        return values.getOrDefault(flag, defaultValue);
    }

    public int getInt(String flag, int defaultValue) {
        String value = values.get(flag);
        if (value == null) return defaultValue;
        try {
            return Integer.parseInt(value);
        } catch (NumberFormatException e) {
            System.err.println("Valor inválido para " + flag + ": " + value + " (esperado inteiro)");
            System.exit(1);
            return defaultValue;
        }
    }
}
