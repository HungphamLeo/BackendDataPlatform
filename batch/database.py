"""MySQL database connection and helpers for data-platform-batch."""
from sqlalchemy import create_engine, text
from sqlalchemy.orm import DeclarativeBase, sessionmaker, Session
import config

_engine = create_engine(config.MYSQL_DSN, pool_pre_ping=True, pool_size=5)
SessionLocal = sessionmaker(bind=_engine, autoflush=False, autocommit=False)


class Base(DeclarativeBase):
    pass


def get_session() -> Session:
    """Return a new SQLAlchemy session.  Caller is responsible for closing it."""
    return SessionLocal()


def init_db() -> None:
    """Create all tables defined via ORM models if they don't already exist."""
    from models import HistoricalOHLCV, IndicatorResult  # noqa: F401 — ensure models are registered
    Base.metadata.create_all(_engine)
