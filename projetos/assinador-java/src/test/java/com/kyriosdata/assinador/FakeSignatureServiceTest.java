package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import org.junit.jupiter.api.Test;

import java.util.Base64;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class FakeSignatureServiceTest {

    private final FakeSignatureService service = new FakeSignatureService();

    // ─── sign — cenários de sucesso ───────────────────────────────────────────

    @Test
    void signDeveRetornarAssinaturaSimuladaParaConteudoBase64Valido() {
        SignRequest request = new SignRequest();
        request.setContent(Base64.getEncoder().encodeToString("conteudo de teste".getBytes()));

        SignatureResponse response = service.sign(request);

        assertNotNull(response);
        assertTrue(response.isValid());
        assertEquals(FakeSignatureService.FAKE_SIGNATURE, response.getSignature());
        assertEquals("Assinatura criada com sucesso", response.getMessage());
    }

    @Test
    void signDeveAceitarTokenOpcional() {
        SignRequest request = new SignRequest();
        request.setContent(Base64.getEncoder().encodeToString("dados".getBytes()));
        request.setToken("meu-token-pin");

        SignatureResponse response = service.sign(request);

        assertTrue(response.isValid());
    }

    // ─── sign — cenários de erro ──────────────────────────────────────────────

    @Test
    void signDeveRejeitarContentNulo() {
        SignRequest request = new SignRequest();
        request.setContent(null);

        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertNull(response.getSignature());
        assertTrue(response.getMessage().contains("content"));
    }

    @Test
    void signDeveRejeitarContentVazio() {
        SignRequest request = new SignRequest();
        request.setContent("   ");

        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("content"));
    }

    @Test
    void signDeveRejeitarContentQueNaoEBase64() {
        SignRequest request = new SignRequest();
        request.setContent("isso não é base64 @@##!!");

        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("Base64"));
    }

    @Test
    void signDeveRejeitarTokenVazioQuandoInformado() {
        SignRequest request = new SignRequest();
        request.setContent(Base64.getEncoder().encodeToString("x".getBytes()));
        request.setToken("   ");

        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("token"));
    }

    @Test
    void signDeveRejeitarRequestNulo() {
        SignatureResponse response = service.sign(null);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("nula"));
    }

    // ─── validate — cenários de sucesso ──────────────────────────────────────

    @Test
    void validateDeveRetornarValidoParaAssinaturaCorreta() {
        ValidateRequest request = new ValidateRequest();
        request.setContent(Base64.getEncoder().encodeToString("conteudo".getBytes()));
        request.setSignature(FakeSignatureService.FAKE_SIGNATURE);

        SignatureResponse response = service.validate(request);

        assertNotNull(response);
        assertTrue(response.isValid());
        assertEquals("Assinatura válida", response.getMessage());
    }

    // ─── validate — cenários de erro ─────────────────────────────────────────

    @Test
    void validateDeveRetornarInvalidoParaAssinaturaErrada() {
        ValidateRequest request = new ValidateRequest();
        request.setContent(Base64.getEncoder().encodeToString("conteudo".getBytes()));
        request.setSignature("ASSINATURA_INCORRETA_==");

        SignatureResponse response = service.validate(request);

        assertNotNull(response);
        assertFalse(response.isValid());
        assertEquals("Assinatura inválida", response.getMessage());
    }

    @Test
    void validateDeveRejeitarContentNulo() {
        ValidateRequest request = new ValidateRequest();
        request.setContent(null);
        request.setSignature(FakeSignatureService.FAKE_SIGNATURE);

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("content"));
    }

    @Test
    void validateDeveRejeitarSignatureNula() {
        ValidateRequest request = new ValidateRequest();
        request.setContent(Base64.getEncoder().encodeToString("x".getBytes()));
        request.setSignature(null);

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("signature"));
    }

    @Test
    void validateDeveRejeitarSignatureVazia() {
        ValidateRequest request = new ValidateRequest();
        request.setContent(Base64.getEncoder().encodeToString("x".getBytes()));
        request.setSignature("");

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("signature"));
    }

    @Test
    void validateDeveRejeitarRequestNulo() {
        SignatureResponse response = service.validate(null);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("nula"));
    }

    // ─── validateSignRequest direto ───────────────────────────────────────────

    @Test
    void validateSignRequestDeveRetornarListaVaziaParaRequestValido() {
        SignRequest request = new SignRequest();
        request.setContent(Base64.getEncoder().encodeToString("payload".getBytes()));

        List<String> errors = service.validateSignRequest(request);

        assertTrue(errors.isEmpty(), "Não deveria haver erros: " + errors);
    }

    @Test
    void validateValidateRequestDeveRetornarListaVaziaParaRequestValido() {
        ValidateRequest request = new ValidateRequest();
        request.setContent(Base64.getEncoder().encodeToString("payload".getBytes()));
        request.setSignature(Base64.getEncoder().encodeToString("sig".getBytes()));

        List<String> errors = service.validateValidateRequest(request);

        assertTrue(errors.isEmpty(), "Não deveria haver erros: " + errors);
    }
}
