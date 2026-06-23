package com.kyriosdata.assinador.pkcs11;

import com.kyriosdata.assinador.FakeSignatureService;
import com.kyriosdata.assinador.SignatureService;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

import java.security.Provider;
import java.security.Security;
import java.util.List;
import java.util.logging.Logger;

/**
 * Implementação de {@link SignatureService} com suporte a PKCS#11 via {@code SunPKCS11}.
 *
 * <p>Quando um dispositivo PKCS#11 (token ou smart card) está disponível, usa-o para
 * operações de assinatura. Caso contrário, delega para {@link FakeSignatureService} com aviso.
 *
 * <p>Para uso com SoftHSM2 em testes de integração, configure o arquivo de configuração PKCS#11:
 * <pre>
 *   --pkcs11-config /path/to/softhsm2.cfg
 * </pre>
 *
 * <p>Arquivo de configuração de exemplo ({@code softhsm2.cfg}):
 * <pre>
 *   name = SoftHSM2
 *   library = /usr/lib/softhsm/libsofthsm2.so
 *   slot = 0
 * </pre>
 */
public class Pkcs11SignatureService implements SignatureService {

    private static final Logger LOG = Logger.getLogger(Pkcs11SignatureService.class.getName());

    private final SignatureService delegate;
    private final boolean pkcs11Available;

    public Pkcs11SignatureService(String pkcs11ConfigPath) {
        boolean available = false;
        SignatureService svc = new FakeSignatureService();

        if (pkcs11ConfigPath != null && !pkcs11ConfigPath.isBlank()) {
            try {
                Provider pkcs11Provider = Security.getProvider("SunPKCS11");
                if (pkcs11Provider != null) {
                    Provider configured = pkcs11Provider.configure(pkcs11ConfigPath);
                    Security.addProvider(configured);
                    LOG.info("SunPKCS11 configurado: " + pkcs11ConfigPath);
                    available = true;
                    svc = new Pkcs11DelegateService(configured);
                } else {
                    LOG.warning("Provider SunPKCS11 não disponível neste JDK.");
                }
            } catch (Exception e) {
                LOG.warning("Falha ao configurar PKCS#11 (usando simulação): " + e.getMessage());
            }
        } else {
            LOG.info("Nenhuma configuração PKCS#11 fornecida; usando simulação.");
        }

        this.pkcs11Available = available;
        this.delegate = svc;
    }

    @Override
    public SignatureResponse sign(SignRequest request) {
        if (!pkcs11Available) {
            LOG.warning("Dispositivo PKCS#11 não disponível; retornando assinatura simulada.");
        }
        return delegate.sign(request);
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        if (!pkcs11Available) {
            LOG.warning("Dispositivo PKCS#11 não disponível; retornando validação simulada.");
        }
        return delegate.validate(request);
    }

    public boolean isPkcs11Available() {
        return pkcs11Available;
    }

    /**
     * Delega operações para o provider PKCS#11 configurado.
     * Neste protótipo, simula a operação usando o provider como chave de contexto.
     */
    private static class Pkcs11DelegateService extends FakeSignatureService {

        private final Provider provider;

        Pkcs11DelegateService(Provider provider) {
            this.provider = provider;
        }

        @Override
        public SignatureResponse sign(SignRequest request) {
            List<String> errors = validateSignRequest(request);
            if (!errors.isEmpty()) {
                return SignatureResponse.error(String.join("; ", errors));
            }
            // Em produção, aqui seria: KeyStore ks = KeyStore.getInstance("PKCS11", provider); ...
            String pkcs11Sig = "PKCS11_SIM_" + provider.getName() + "_BASE64==";
            return SignatureResponse.success(pkcs11Sig, "Assinatura criada via PKCS#11 simulado");
        }
    }
}
