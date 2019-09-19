# Secrets Management in K8S (EKS / GKE)

Raw k8s secrets work well for smaller scale deployments and small teams due to out of box experience, but for any large enterprises they will need something with better control. This RFC explores requirements and alternatives for *additional* secrets management mechanisms in K8S.

## Requirements

- **Confidentiality** for secrets, keys, and credentials.
- **Recoverability:** provide mechanisms for key rotation in case of compromise.
- **Auditability:** keep track of what systems and users access confidential data

- **Understandability:** well documented with encryption mechanism and associated components well understandable - reduces chances of misconfiguration and creating seams with exposed vulnerabilities. 
- **Maintainability:** well-supported and active community / user base 
- **Developer-friendliness:** support *DevOps processes* - it is important that the tool can make the processes associated with working with secrets (generation, rotation, revocation, assignment, and sharing) easy and organized.
- **Portability:** portable across (Gruntwork) supported k8s engines - potentially beyond those



## Options:

### KMS (AWS KMS / Cloud KMS):

#### How:

TBD

#### Pros:

* Quick to implement
* Stable and well understood technology
* ...

#### Cons:

* TBD



### Vault

https://www.vaultproject.io/

#### How:

TBD

#### Pros:

* Stack Agnostic: You can use Vault secrets in applications running on any platform / cloud provider
* Battle-tested: Widely adopted and used 
* Rich feature set

#### Cons:

* Complexity
* Increased infrastructure cost



### Sealed Secrets

https://github.com/bitnami-labs/sealed-secrets

#### How:

We'd write the contents of the cache to the local filesystem and would read it into memory when the app loads/update the file contents when the cache is updated. Likely this solution would be paired with keeping the cache in memory and we'd use the local disk storage to allow app-restarts without destroying the cache.

#### Pros:

- Fast
- Simple
- No added infra
- No added costs

#### Cons:

- Cache per sever means that they may not all be in sync (depending on how we implement the update logic)



### SOPS

https://github.com/mozilla/sops

#### How:

#### Pros:

- 

#### Cons:

- 



### Kamus

https://github.com/Soluto/kamus

#### How:

#### Pros:

- 

#### Cons:

- 

