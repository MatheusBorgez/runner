package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

import java.util.ArrayList;
import java.util.Base64;
import java.util.List;

/**
 * Implementação simulada de {@link SignatureService}.
 * Valida parâmetros conforme as especificações FHIR Goiás e retorna
 * respostas pré-construídas para requisições válidas.
 */
public class FakeSignatureService implements SignatureService {

    static final String FAKE_SIGNATURE = "MOCKED_SIGNATURE_BASE64_==";

    @Override
    public SignatureResponse sign(SignRequest request) {
        List<String> errors = validateSignRequest(request);
        if (!errors.isEmpty()) {
            return SignatureResponse.error(String.join("; ", errors));
        }
        return SignatureResponse.success(FAKE_SIGNATURE, "Assinatura criada com sucesso");
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        List<String> errors = validateValidateRequest(request);
        if (!errors.isEmpty()) {
            return SignatureResponse.error(String.join("; ", errors));
        }
        boolean isValid = FAKE_SIGNATURE.equals(request.getSignature());
        return new SignatureResponse(request.getSignature(), isValid, isValid ? "Assinatura válida" : "Assinatura inválida");
    }

    protected List<String> validateSignRequest(SignRequest request) {
        List<String> errors = new ArrayList<>();
        if (request == null) {
            errors.add("Requisição nula");
            return errors;
        }
        if (isBlank(request.getContent())) {
            errors.add("'content' é obrigatório e não pode ser vazio");
        } else {
            validateBase64Field("content", request.getContent(), errors);
        }
        if (request.getToken() != null && request.getToken().isBlank()) {
            errors.add("'token' não pode ser string vazia quando informado; omita o campo se não aplicável");
        }
        return errors;
    }

    protected List<String> validateValidateRequest(ValidateRequest request) {
        List<String> errors = new ArrayList<>();
        if (request == null) {
            errors.add("Requisição nula");
            return errors;
        }
        if (isBlank(request.getContent())) {
            errors.add("'content' é obrigatório e não pode ser vazio");
        } else {
            validateBase64Field("content", request.getContent(), errors);
        }
        if (isBlank(request.getSignature())) {
            errors.add("'signature' é obrigatório e não pode ser vazio");
        } else {
            validateBase64Field("signature", request.getSignature(), errors);
        }
        return errors;
    }

    private void validateBase64Field(String fieldName, String value, List<String> errors) {
        try {
            byte[] decoded = Base64.getDecoder().decode(value.trim());
            if (decoded.length == 0) {
                errors.add("'" + fieldName + "' decodificou para conteúdo vazio");
            }
        } catch (IllegalArgumentException e) {
            errors.add("'" + fieldName + "' não é Base64 válido: " + e.getMessage());
        }
    }

    private boolean isBlank(String value) {
        return value == null || value.isBlank();
    }
}
