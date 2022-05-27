import useUser from "@aqueducthq/common/src/components/hooks/useUser";
import DataPage from "@aqueducthq/common/src/components/pages/data";
import { useRouter } from "next/router";
import React from "react";

export { getServerSideProps } from "@aqueducthq/common/src/components/pages/getServerSideProps";

const Data: React.FC = () => {
  const { user, loading, success } = useUser();
  if (loading) {
    return null;
  }

  const router = useRouter();
  if (!success) {
    router.push("/login");
    return null;
  }

  return <DataPage user={user} />;
};

export default Data;
