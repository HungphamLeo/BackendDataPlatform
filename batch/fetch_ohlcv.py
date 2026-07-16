"""Fetch historical OHLCV data from Binance REST API and persist to MySQL."""
import logging
from datetime import datetime, timedelta, timezone

import requests
import pandas as pd
from sqlalchemy.dialects.mysql import insert

import config
from database import get_session
from models import HistoricalOHLCV

logger = logging.getLogger(__name__)

_BINANCE_KLINES_URL = f"{config.BINANCE_BASE_URL}/api/v3/klines"
_INTERVAL_MAP = {"1m": "1m", "5m": "5m", "1h": "1h", "1d": "1d"}


def _fetch_klines(symbol: str, interval: str, start_ms: int, end_ms: int) -> list[list]:
    """Fetch raw klines from Binance, paginating if necessary."""
    all_klines: list = []
    limit = 1000

    while start_ms < end_ms:
        params = {
            "symbol": symbol,
            "interval": interval,
            "startTime": start_ms,
            "endTime": end_ms,
            "limit": limit,
        }
        resp = requests.get(_BINANCE_KLINES_URL, params=params, timeout=10)
        resp.raise_for_status()
        batch = resp.json()
        if not batch:
            break
        all_klines.extend(batch)
        # next page: start from the last candle's close_time + 1 ms
        start_ms = batch[-1][6] + 1
        if len(batch) < limit:
            break

    return all_klines


def _klines_to_df(raw: list[list], symbol: str, exchange: str, interval: str) -> pd.DataFrame:
    """Convert raw Binance kline response to a normalised DataFrame."""
    if not raw:
        return pd.DataFrame()
    df = pd.DataFrame(raw, columns=[
        "open_time", "open", "high", "low", "close", "volume",
        "close_time", "quote_volume", "num_trades",
        "taker_buy_base", "taker_buy_quote", "ignore",
    ])
    df = df[["open_time", "open", "high", "low", "close", "volume"]].copy()
    df["timestamp"] = pd.to_datetime(df["open_time"], unit="ms", utc=True).dt.tz_localize(None)
    df["symbol"] = symbol
    df["exchange"] = exchange
    df["interval"] = interval
    for col in ("open", "high", "low", "close", "volume"):
        df[col] = df[col].astype(float)
    return df[["symbol", "exchange", "interval", "open", "high", "low", "close", "volume", "timestamp"]]


def fetch_and_persist(symbol: str, interval: str, lookback_days: int = None) -> int:
    """Fetch OHLCV for *symbol*+*interval* and upsert to DB. Returns rows upserted."""
    lookback_days = lookback_days or config.LOOKBACK_DAYS
    end_dt = datetime.now(tz=timezone.utc)
    start_dt = end_dt - timedelta(days=lookback_days)

    raw = _fetch_klines(
        symbol=symbol,
        interval=_INTERVAL_MAP.get(interval, interval),
        start_ms=int(start_dt.timestamp() * 1000),
        end_ms=int(end_dt.timestamp() * 1000),
    )
    df = _klines_to_df(raw, symbol=symbol, exchange="binance", interval=interval)
    if df.empty:
        return 0

    session = get_session()
    try:
        rows = df.to_dict(orient="records")
        stmt = insert(HistoricalOHLCV).values(rows)
        stmt = stmt.on_duplicate_key_update(
            open=stmt.inserted.open,
            high=stmt.inserted.high,
            low=stmt.inserted.low,
            close=stmt.inserted.close,
            volume=stmt.inserted.volume,
        )
        session.execute(stmt)
        session.commit()
        return len(rows)
    except Exception as exc:
        session.rollback()
        logger.error("persist ohlcv failed: symbol=%s interval=%s error=%s", symbol, interval, exc)
        raise
    finally:
        session.close()


def run_all() -> None:
    """Fetch OHLCV for all configured symbols and intervals."""
    for symbol in config.SYMBOLS:
        for interval in config.INTERVALS:
            try:
                n = fetch_and_persist(symbol, interval)
                logger.info("fetched ohlcv symbol=%s interval=%s rows=%d", symbol, interval, n)
            except Exception as exc:
                logger.error("fetch_ohlcv failed symbol=%s interval=%s: %s", symbol, interval, exc)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    run_all()
