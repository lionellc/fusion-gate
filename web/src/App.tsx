import { Button, Toast } from "@douyinfe/semi-ui";

function App() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <Button onClick={() => Toast.success({ content: "welcome" })}>
        Hello FusionGate
      </Button>
    </div>
  );
}

export default App;
