"use client";

import { useEffect, useRef, useCallback } from "react";
import * as d3 from "d3";
import type { BubbleData } from "@/lib/api";

interface BubbleChartProps {
  data: BubbleData[];
  width?: number;
  height?: number;
  onBubbleClick?: (item: BubbleData) => void;
}

interface SimNode extends d3.SimulationNodeDatum, BubbleData {
  r: number;
}

export function BubbleChart({
  data,
  width = 800,
  height = 500,
  onBubbleClick,
}: BubbleChartProps) {
  const svgRef = useRef<SVGSVGElement>(null);

  const render = useCallback(() => {
    if (!svgRef.current || data.length === 0) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll("*").remove();

    const maxVal = d3.max(data, (d) => d.value) || 1;
    const radiusScale = d3
      .scaleSqrt()
      .domain([0, maxVal])
      .range([15, Math.min(width, height) / 6]);

    const nodes: SimNode[] = data.map((d) => ({
      ...d,
      r: radiusScale(d.value),
      x: width / 2 + (Math.random() - 0.5) * 100,
      y: height / 2 + (Math.random() - 0.5) * 100,
    }));

    const simulation = d3
      .forceSimulation(nodes)
      .force("center", d3.forceCenter(width / 2, height / 2))
      .force("charge", d3.forceManyBody().strength(5))
      .force(
        "collide",
        d3
          .forceCollide<SimNode>()
          .radius((d) => d.r + 2)
          .strength(0.9)
          .iterations(3)
      )
      .force("x", d3.forceX(width / 2).strength(0.05))
      .force("y", d3.forceY(height / 2).strength(0.05));

    const g = svg
      .append("g")
      .attr("class", "bubbles");

    // Zoom
    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.5, 3])
      .on("zoom", (event) => {
        g.attr("transform", event.transform);
      });
    svg.call(zoom);

    const bubbleGroups = g
      .selectAll("g.bubble")
      .data(nodes)
      .enter()
      .append("g")
      .attr("class", "bubble")
      .style("cursor", "pointer")
      .on("click", (_, d) => onBubbleClick?.(d))
      .call(
        d3
          .drag<SVGGElement, SimNode>()
          .on("start", (event, d) => {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on("drag", (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on("end", (event, d) => {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          })
      );

    bubbleGroups
      .append("circle")
      .attr("r", (d) => d.r)
      .attr("fill", (d) => d.color)
      .attr("fill-opacity", 0.75)
      .attr("stroke", (d) => d.color)
      .attr("stroke-width", 1.5)
      .attr("stroke-opacity", 0.9);

    bubbleGroups
      .append("text")
      .text((d) => d.label)
      .attr("text-anchor", "middle")
      .attr("dy", "-0.2em")
      .attr("fill", "white")
      .attr("font-size", (d) => Math.max(9, Math.min(d.r / 3, 14)))
      .attr("font-weight", 600)
      .attr("pointer-events", "none");

    bubbleGroups
      .append("text")
      .text((d) => String(d.value))
      .attr("text-anchor", "middle")
      .attr("dy", "1em")
      .attr("fill", "white")
      .attr("fill-opacity", 0.7)
      .attr("font-size", (d) => Math.max(8, Math.min(d.r / 4, 12)))
      .attr("pointer-events", "none");

    simulation.on("tick", () => {
      bubbleGroups.attr("transform", (d) => `translate(${d.x},${d.y})`);
    });

    return () => {
      simulation.stop();
    };
  }, [data, width, height, onBubbleClick]);

  useEffect(() => {
    const cleanup = render();
    return () => cleanup?.();
  }, [render]);

  if (data.length === 0) {
    return (
      <div className="flex h-[400px] items-center justify-center text-muted-foreground">
        No data available for bubble chart.
      </div>
    );
  }

  return (
    <svg
      ref={svgRef}
      width="100%"
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className="rounded-xl border border-border/50 bg-card/30"
    />
  );
}
