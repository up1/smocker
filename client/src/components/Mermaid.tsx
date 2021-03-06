import { Alert } from "antd";
import mermaidAPI from "mermaid";
import * as React from "react";
import "./Mermaid.scss";

export const Mermaid = ({
  name,
  chart,
  onChange,
}: {
  name: string;
  chart: string;
  onChange?: (svg: string) => unknown;
}): JSX.Element => {
  const [diagram, setDiagram] = React.useState("");
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    const cb = (svg = "") => {
      setDiagram(svg);
      setError("");
      onChange && onChange(svg);
    };
    try {
      mermaidAPI.parse(chart);
      mermaidAPI.initialize({ startOnLoad: false });
      mermaidAPI.render(name, chart, cb);
    } catch (e) {
      setDiagram("");
      console.error(e);
      setError(e.str || `${e}`);
    }
  }, [name, chart]);

  return (
    <div className="mermaid">
      <div dangerouslySetInnerHTML={{ __html: diagram }} />
      {error && (
        <Alert
          message="Unable to render"
          description={error}
          type="error"
          showIcon
        />
      )}
    </div>
  );
};
