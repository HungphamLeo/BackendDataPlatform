"""Compute technical indicators from historical OHLCV and persist results."""
import json
import logging
from datetime import datetime

import numpy as np
import pandas as pd
from sqlalchemy import select, and_

import config
from database import get_session
from models import HistoricalOHLCV, IndicatorResult

logger = logging.getLogger(__name__)


# ── indicator computation helpers ──────────────────────────────────────────

def compute_sma(close: pd.Series, period: int = 20) -> pd.Series:
    return close.rolling(window=period, min_periods=period).mean()


def compute_ema(close: pd.Series, period: int = 20) -> pd.Series:
    return close.ewm(span=period, min_periods=period, adjust=False).mean()


def compute_rsi(close: pd.Series, period: int = 14) -> pd.Series:
    delta = close.diff()
    gain = delta.clip(lower=0)
    loss = -delta.clip(upper=0)
    avg_gain = gain.rolling(window=period, min_periods=period).mean()
    avg_loss = loss.rolling(window=period, min_periods=period).mean()
    rs = avg_gain / avg_loss.replace(0, np.nan)
    return 100 - (100 / (1 + rs))


def compute_atr(high: pd.Series, low: pd.Series, close: pd.Series, period: int = 14) -> pd.Series:
    prev_close = close.shift(1)
    tr = pd.concat([
        high - low,
        (high - prev_close).abs(),
        (low - prev_close).abs(),
    ], axis=1).max(axis=1)
    return tr.rolling(window=period, min_periods=period).mean()


def compute_bbands(close: pd.Series, period: int = 20, num_std: float = 2.0):
    """Returns (middle, upper, lower) Series."""
    middle = compute_sma(close, period)
    std = close.rolling(window=period, min_periods=period).std()
    return middle, middle + num_std * std, middle - num_std * std


# ── main pipeline ──────────────────────────────────────────────────────────

def _load_ohlcv(session, symbol: str, exchange: str, interval: str) -> pd.DataFrame:
    rows = session.execute(
        select(HistoricalOHLCV).where(
            and_(
                HistoricalOHLCV.symbol == symbol,
                HistoricalOHLCV.exchange == exchange,
                HistoricalOHLCV.interval == interval,
            )
        ).order_by(HistoricalOHLCV.timestamp)
    ).scalars().all()

    if not rows:
        return pd.DataFrame()

    return pd.DataFrame([{
        "timestamp": r.timestamp,
        "open": r.open, "high": r.high, "low": r.low,
        "close": r.close, "volume": r.volume,
    } for r in rows])


def _persist_indicator_rows(session, rows: list[dict]) -> None:
    if not rows:
        return
    session.bulk_insert_mappings(IndicatorResult, rows)  # type: ignore[arg-type]
    session.commit()


def compute_and_persist(symbol: str, exchange: str, interval: str) -> int:
    """Compute all indicators for symbol/exchange/interval and persist. Returns rows written."""
    session = get_session()
    try:
        df = _load_ohlcv(session, symbol, exchange, interval)
        if df.empty:
            logger.warning("no ohlcv data for %s/%s/%s — skipping", symbol, exchange, interval)
            return 0

        close = df["close"]
        high = df["high"]
        low = df["low"]
        timestamps = df["timestamp"]

        indicator_rows: list[dict] = []

        def _add(name: str, series: pd.Series, meta: dict | None = None):
            for ts, val in zip(timestamps, series):
                if pd.isna(val):
                    continue
                indicator_rows.append({
                    "symbol": symbol,
                    "exchange": exchange,
                    "interval": interval,
                    "indicator_name": name,
                    "value": float(val),
                    "meta_json": json.dumps(meta) if meta else None,
                    "timestamp": ts,
                    "created_at": datetime.utcnow(),
                })

        _add("SMA_20", compute_sma(close, 20))
        _add("EMA_20", compute_ema(close, 20))
        _add("RSI_14", compute_rsi(close, 14))
        _add("ATR_14", compute_atr(high, low, close, 14))

        mid, upper, lower = compute_bbands(close, 20, 2.0)
        for ts, m, u, l in zip(timestamps, mid, upper, lower):
            if pd.isna(m):
                continue
            indicator_rows.append({
                "symbol": symbol, "exchange": exchange, "interval": interval,
                "indicator_name": "BBANDS",
                "value": float(m),
                "meta_json": json.dumps({"upper": float(u), "lower": float(l)}),
                "timestamp": ts,
                "created_at": datetime.utcnow(),
            })

        # Delete stale indicator data before inserting fresh
        session.query(IndicatorResult).filter(
            and_(
                IndicatorResult.symbol == symbol,
                IndicatorResult.exchange == exchange,
                IndicatorResult.interval == interval,
            )
        ).delete()

        _persist_indicator_rows(session, indicator_rows)
        return len(indicator_rows)

    except Exception as exc:
        session.rollback()
        logger.error("compute_and_persist failed %s/%s/%s: %s", symbol, exchange, interval, exc)
        raise
    finally:
        session.close()


def run_all() -> None:
    for symbol in config.SYMBOLS:
        for interval in config.INTERVALS:
            try:
                n = compute_and_persist(symbol, "binance", interval)
                logger.info("computed indicators symbol=%s interval=%s rows=%d", symbol, interval, n)
            except Exception as exc:
                logger.error("compute_indicators failed %s/%s: %s", symbol, interval, exc)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
    run_all()
