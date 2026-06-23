package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

import java.util.ArrayList;
import java.util.Base64;
import java.util.List;

/**
 * Implementação simulada de {@link SignatureService}.
 *
 * <p>Realiza validação rigorosa dos parâmetros conforme as especificações
 * do caso de uso de assinatura digital da Plataforma HubSaúde (FHIR Goiás),
 * retornando respostas pré-construídas para requisições válidas.
 *
 * <p>O foco está na validação de parâmetros. Nenhuma criptografia real é executada.
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
        String message = isValid ? "Assinatura válida" : "Assinatura inválida";
        return new SignatureResponse(request.getSignature(), isValid, message);
    }

    /**
     * Valida todos os campos obrigatórios e o formato da requisição de criação.
     *
     * @return lista de mensagens de erro; vazia se válido
     */
    protected List<String> validateSignRequest(SignRequest request) {
        List<String> errors = new ArrayList<>();

        if (request == null) {
            errors.add("Requisição nula");
            return errors;
        }

        // content: obrigatório, não vazio, máx 10 MB em Base64
        if (isBlank(request.getContent())) {
            errors.add("'content' é obrigatório e não pode ser vazio");
        } else {
            validateBase64Field("content", request.getContent(), errors);
        }

        // token: opcional, mas se presente não pode ser vazio
        if (request.getToken() != null && request.getToken().isBlank()) {
            errors.add("'token' não pode ser string vazia quando informado; omita o campo se não aplicável");
        }

        return errors;
    }

    /**
     * Valida todos os campos obrigatórios e o formato da requisição de validação.
     *
     * @return lista de mensagens de erro; vazia se válido
     */
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
