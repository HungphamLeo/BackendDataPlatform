"""Configuration for the data-platform-batch service.

All settings are read from environment variables (or a .env file).
"""
import os
from dotenv import load_dotenv

load_dotenv()

# --- MySQL ---
MYSQL_DSN: str = os.getenv(
    "MYSQL_DSN",
    "mysql+mysqlconnector://root:password@localhost:3306/data_platform",
)

# --- Binance REST ---
BINANCE_BASE_URL: str = "https://api.binance.com"
BINANCE_FUTURES_BASE_URL: str = "https://fapi.binance.com"

# --- Symbols and intervals to process ---
SYMBOLS: list[str] = os.getenv("SYMBOLS", "BTCUSDT,ETHUSDT,BNBUSDT").split(",")
INTERVALS: list[str] = os.getenv("INTERVALS", "1m,5m,1h,1d").split(",")
LOOKBACK_DAYS: int = int(os.getenv("LOOKBACK_DAYS", "90"))

# --- gRPC server ---
GRPC_PORT: int = int(os.getenv("GRPC_PORT", "9093"))

# --- Schedule ---
FETCH_INTERVAL_MINUTES: int = int(os.getenv("FETCH_INTERVAL_MINUTES", "60"))
COMPUTE_INTERVAL_MINUTES: int = int(os.getenv("COMPUTE_INTERVAL_MINUTES", "65"))
