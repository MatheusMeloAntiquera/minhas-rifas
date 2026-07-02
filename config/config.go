package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MongoURI                  string `envconfig:"MONGO_URI" default:"mongodb://localhost:27017"`
	MongoDatabaseName         string `envconfig:"MONGO_DATABASE_NAME" default:"minhas-rifas"`
	Port                      string `envconfig:"PORT" default:"8080"`
	ClerkSecretKey            string `envconfig:"CLERK_SECRET_KEY" required:"true"`
	ClerkWebhookSigningSecret string `envconfig:"CLERK_WEBHOOK_SIGNING_SECRET" required:"true"`
}

func New() (cfg Config, err error) {
	// Carrega um arquivo .env se existir (desenvolvimento local). Em produção,
	// as variáveis costumam vir do ambiente, então a ausência do arquivo não é erro.
	_ = godotenv.Load()

	err = envconfig.Process("", &cfg)
	return
}
