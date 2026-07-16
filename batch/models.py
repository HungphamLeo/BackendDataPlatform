"""SQLAlchemy ORM models for data-platform-batch."""
from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, Index
from database import Base


class HistoricalOHLCV(Base):
    __tablename__ = "historical_ohlcv"

    id = Column(Integer, primary_key=True, autoincrement=True)
    symbol = Column(String(20), nullable=False)
    exchange = Column(String(20), nullable=False)
    interval = Column(String(10), nullable=False)
    open = Column(Float, nullable=False)
    high = Column(Float, nullable=False)
    low = Column(Float, nullable=False)
    close = Column(Float, nullable=False)
    volume = Column(Float, nullable=False)
    timestamp = Column(DateTime, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    __table_args__ = (
        Index("idx_ohlcv_sym_exch_intv_ts", "symbol", "exchange", "interval", "timestamp", unique=True),
    )


class IndicatorResult(Base):
    __tablename__ = "indicator_result"

    id = Column(Integer, primary_key=True, autoincrement=True)
    symbol = Column(String(20), nullable=False)
    exchange = Column(String(20), nullable=False)
    interval = Column(String(10), nullable=False)
    indicator_name = Column(String(20), nullable=False)
    value = Column(Float, nullable=False)
    meta_json = Column(String(512), nullable=True)  # for e.g. Bollinger upper/lower
    timestamp = Column(DateTime, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    __table_args__ = (
        Index("idx_ind_sym_exch_intv_name_ts", "symbol", "exchange", "interval", "indicator_name", "timestamp"),
    )
