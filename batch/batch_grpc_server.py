"""gRPC server implementing the BatchData service.

The .proto-generated Python stubs (batch_data_pb2, batch_data_pb2_grpc) are
expected to be generated via `python -m grpc_tools.protoc` — see README for
the exact command.  Until then, this module defines the servicer class and
`serve_grpc` helper so run_all.py can wire it in.
"""
from __future__ import annotations

import logging
from concurrent import futures
from datetime import datetime

import grpc

# generated stubs — run `make proto-py` to create these
try:
    import batch_data_pb2 as pb2  # type: ignore
    import batch_data_pb2_grpc as pb2_grpc  # type: ignore
    _STUBS_AVAILABLE = True
except ImportError:
    _STUBS_AVAILABLE = False

import config
from database import get_session
from models import HistoricalOHLCV, IndicatorResult
from sqlalchemy import and_, select

logger = logging.getLogger(__name__)


class BatchDataServicer:
    """Implements the BatchData gRPC service."""

    def GetOHLCV(self, request, context):
        session = get_session()
        try:
            q = select(HistoricalOHLCV).where(
                and_(
                    HistoricalOHLCV.symbol == request.symbol,
                    HistoricalOHLCV.exchange == request.exchange,
                    HistoricalOHLCV.interval == request.interval,
                )
            ).order_by(HistoricalOHLCV.timestamp)

            if request.HasField("from"):
                q = q.where(HistoricalOHLCV.timestamp >= request.from_.ToDatetime())
            if request.HasField("to"):
                q = q.where(HistoricalOHLCV.timestamp <= request.to.ToDatetime())
            if request.limit > 0:
                q = q.limit(request.limit)

            rows = session.execute(q).scalars().all()
            bars = []
            for r in rows:
                bar = pb2.OHLCVBar(
                    symbol=r.symbol, exchange=r.exchange, interval=r.interval,
                    open=r.open, high=r.high, low=r.low, close=r.close, volume=r.volume,
                )
                bar.timestamp.FromDatetime(r.timestamp)
                bars.append(bar)
            return pb2.OHLCVResponse(bars=bars)
        finally:
            session.close()

    def GetIndicator(self, request, context):
        session = get_session()
        try:
            q = select(IndicatorResult).where(
                and_(
                    IndicatorResult.symbol == request.symbol,
                    IndicatorResult.exchange == request.exchange,
                    IndicatorResult.interval == request.interval,
                    IndicatorResult.indicator_name == request.indicator_name,
                )
            ).order_by(IndicatorResult.timestamp)

            if request.HasField("from"):
                q = q.where(IndicatorResult.timestamp >= request.from_.ToDatetime())
            if request.HasField("to"):
                q = q.where(IndicatorResult.timestamp <= request.to.ToDatetime())
            if request.limit > 0:
                q = q.limit(request.limit)

            rows = session.execute(q).scalars().all()
            points = []
            for r in rows:
                pt = pb2.IndicatorPoint(
                    symbol=r.symbol, exchange=r.exchange, interval=r.interval,
                    indicator_name=r.indicator_name, value=r.value,
                    meta_json=r.meta_json or "",
                )
                pt.timestamp.FromDatetime(r.timestamp)
                points.append(pt)
            return pb2.IndicatorResponse(points=points)
        finally:
            session.close()

    def ListAvailableSymbols(self, request, context):
        session = get_session()
        try:
            q = select(HistoricalOHLCV.symbol).distinct()
            if request.exchange:
                q = q.where(HistoricalOHLCV.exchange == request.exchange)
            rows = session.execute(q).scalars().all()
            return pb2.ListSymbolsResponse(symbols=list(rows))
        finally:
            session.close()


def serve_grpc(port: int = 9093) -> None:
    """Start the gRPC server (blocking)."""
    if not _STUBS_AVAILABLE:
        logger.warning("batch_data_pb2 stubs not found — gRPC server disabled. Run `make proto-py`.")
        return

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pb2_grpc.add_BatchDataServicer_to_server(BatchDataServicer(), server)
    server.add_insecure_port(f"[::]:{port}")
    server.start()
    logger.info("BatchData gRPC server listening on :%d", port)
    server.wait_for_termination()
